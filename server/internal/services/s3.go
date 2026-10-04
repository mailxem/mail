package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"kori/internal/utils/logger"

	"kori/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// Ensure S3Service implements FileURLGenerator
var _ models.FileURLGenerator = (*S3Service)(nil)

type S3Service struct {
	client      *s3.Client
	presigner   *s3.Client
	disableACL  bool
	bucketName  string
	endpoint    string
	endpointURL string
	pathStyle   bool
	region      string
	logger      *logger.Logger
	accessKey   string
	secretKey   string
}

func NewS3Service(bucketName, endpoint, region, accessKey, secretKey string) (*S3Service, error) {
	log := logger.New("s3_service")

	// Validate required credentials
	if accessKey == "" || secretKey == "" {
		return nil, log.Error("S3 credentials are empty ❌", fmt.Errorf("accessKey or secretKey is empty"))
	}

	// Preserve legacy endpoint/signing behavior unless the operator opts into a
	// full endpoint URL. This supports private S3-compatible Swarm services.
	endpointURL, signingRegion, pathStyle, err := resolveS3Endpoint(endpoint, region, os.Getenv("S3_ENDPOINT_URL"))
	if err != nil {
		return nil, err
	}

	publicEndpointURL, err := resolveS3PublicEndpoint(os.Getenv("S3_PUBLIC_ENDPOINT_URL"), os.Getenv("S3_ENDPOINT_URL"), signingRegion)
	if err != nil {
		return nil, err
	}

	// Create AWS config with explicit credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(signingRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"", // Session token (not needed for basic auth)
		)),
		config.WithRetryMode(aws.RetryModeStandard),
		config.WithRetryMaxAttempts(3),
	)
	if err != nil {
		return nil, log.Error("Unable to load SDK config ❌", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpointURL)
		o.UsePathStyle = pathStyle
	})

	// Sign browser reads for the public proxy without sending uploads through it.
	presigner := client
	if publicEndpointURL != "" {
		presigner = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(publicEndpointURL)
			o.UsePathStyle = true
		})
	}

	// Limit storage initialization even when the endpoint is unavailable.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = verifyS3Bucket(ctx, client, bucketName, os.Getenv("S3_CREATE_BUCKET") == "true")
	if err != nil {
		return nil, log.Error("Failed to verify S3 credentials ❌", err)
	}

	log.Success("S3 service initialized successfully ✅")

	if publicEndpointURL != "" {
		endpointURL = publicEndpointURL
	}

	return &S3Service{
		client:      client,
		presigner:   presigner,
		disableACL:  os.Getenv("S3_DISABLE_ACL") == "true",
		bucketName:  bucketName,
		endpoint:    endpoint,
		endpointURL: endpointURL,
		pathStyle:   pathStyle,
		region:      region,
		accessKey:   accessKey,
		secretKey:   secretKey,
		logger:      log,
	}, nil
}

// UploadFile uploads a file to S3 or S3-compatible storage and returns the URL
func (s *S3Service) UploadFile(ctx context.Context, file []byte, filename string, acl types.ObjectCannedACL, contentType string) (string, error) {
	s.logger.Info("📤 Starting file upload: %s", filename)

	// Generate unique filename
	ext := filepath.Ext(filename)

	filename = fmt.Sprintf("%s%s", uuid.New().String(), ext)

	s.logger.Info("🔄 Processing upload for file: %s", filename)

	// Preserve hosted ACL behavior unless the operator uses an ACL-disabled bucket.
	objectACL := acl
	if s.disableACL {
		objectACL = ""
	} else if os.Getenv("STORAGE_PROVIDER") == "r2" {
		objectACL = types.ObjectCannedACLPublicRead
	}

	// Upload to storage
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(file),
		ACL:         objectACL,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", s.logger.Error("Failed to upload file to storage ❌", err)
	}

	// Generate URL based on endpoint configuration
	var url string
	if s.pathStyle {
		url = fmt.Sprintf("%s/%s/%s", s.endpointURL, s.bucketName, filename)
	} else if s.endpoint != "" {
		// Custom endpoint (e.g., MinIO)
		url = fmt.Sprintf("https://%s.%s/%s/%s", s.region, s.endpoint, s.bucketName, filename)
	} else {
		// AWS S3
		url = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, filename)
	}

	s.logger.Success("✅ File uploaded successfully: %s", url)
	return url, nil
}

// GetSignedURL implements FileURLGenerator interface
func (s *S3Service) GetSignedURL(ctx context.Context, path string, duration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.presigner)

	s.logger.Info("🔄 Generating pre-signed URL for path: %s", path)

	presignedURL, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	}, s3.WithPresignExpires(duration))

	if err != nil {
		return "", s.logger.Error("Failed to generate pre-signed URL ❌", err)
	}

	s.logger.Success("✅ Generated pre-signed URL successfully")
	return presignedURL.URL, nil
}

// resolveS3Endpoint keeps the historic region-prefixed endpoint working. An
// explicit URL uses the configured AWS signing region and path-style addressing.
func resolveS3Endpoint(endpoint, region, explicit string) (string, string, bool, error) {
	if explicit == "" {
		return fmt.Sprintf("https://%s.%s", region, endpoint), "apac", false, nil
	}
	parsed, err := url.Parse(explicit)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", "", false, fmt.Errorf("S3_ENDPOINT_URL must be an HTTP(S) origin without credentials, query, or path")
	}
	if region == "" {
		return "", "", false, fmt.Errorf("S3_REGION is required with S3_ENDPOINT_URL")
	}
	return strings.TrimRight(explicit, "/"), region, true, nil
}

// A public origin changes object URLs and signatures, not the private S3 client.
func resolveS3PublicEndpoint(public, private, region string) (string, error) {
	if public == "" {
		return "", nil
	}
	if private == "" {
		return "", fmt.Errorf("S3_PUBLIC_ENDPOINT_URL requires S3_ENDPOINT_URL")
	}
	endpoint, _, _, err := resolveS3Endpoint("", region, public)
	if err != nil || !strings.HasPrefix(endpoint, "https://") {
		return "", fmt.Errorf("S3_PUBLIC_ENDPOINT_URL must be an HTTPS origin without credentials, query, or path")
	}
	return endpoint, nil
}

// Bucket creation is opt-in for MinIO-compatible stores and never changes policy.
// Provision AWS regional buckets separately; no LocationConstraint is sent here.
func verifyS3Bucket(ctx context.Context, client *s3.Client, bucket string, create bool) error {
	list := &s3.ListObjectsV2Input{Bucket: aws.String(bucket), MaxKeys: aws.Int32(1)}
	_, err := client.ListObjectsV2(ctx, list)
	if !create || !isS3Error(err, "NoSuchBucket") {
		return err
	}
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil && !isS3Error(err, "BucketAlreadyOwnedByYou") {
		return err
	}
	// A create race is safe only after the configured credentials can list it.
	_, err = client.ListObjectsV2(ctx, list)
	return err
}

func isS3Error(err error, code string) bool {
	var apiError smithy.APIError
	return errors.As(err, &apiError) && apiError.ErrorCode() == code
}
