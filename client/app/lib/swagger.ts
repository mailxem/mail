import { OpenAPIV3 } from 'openapi-types';

export const createRouteSpec = (
  path: string,
  operations: Partial<Record<'get' | 'post' | 'put' | 'delete' | 'patch', OpenAPIV3.OperationObject>>
): OpenAPIV3.PathItemObject => {
  return Object.entries(operations).reduce((acc, [method, operation]) => {
    acc[method] = {
      ...operation,
      security: [{ bearerAuth: [], ApiKeyAuth: [
        {
          type: 'apiKey',
          in: 'header',
          name: 'x-api-key',
        }
      ] }],
      responses: {
        ...operation.responses,
        401: {
          description: 'Unauthorized',
          content: {
            'application/json': {
              schema: {
                type: 'object',
                properties: {
                  message: { type: 'string' }
                }
              }
            }
          }
        },
        500: {
          description: 'Internal Server Error',
          content: {
            'application/json': {
              schema: {
                type: 'object',
                properties: {
                  message: { type: 'string' }
                }
              }
            }
          }
        }
      }
    };
    return acc;
  }, {} as OpenAPIV3.PathItemObject);
};

export const baseApiSpec: Partial<OpenAPIV3.Document> = {
  openapi: '3.0.0',
  info: {
    title: 'SailMail API',
    version: '1.0.0',
    description: 'API documentation for SailMail platform'
  },
  components: {
    securitySchemes: {
      bearerAuth: {
        type: 'http',
        scheme: 'bearer',
        bearerFormat: 'JWT'
      }
    },
    schemas: {
      Error: {
        type: 'object',
        properties: {
          message: {
            type: 'string'
          }
        }
      }
    }
  },
  security: [
    {
      bearerAuth: []
    }
  ]
}; 