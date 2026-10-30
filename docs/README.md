# Xem API Documentation

This directory contains the comprehensive API documentation for the Xem backend service. The documentation is built using [Mintlify](https://mintlify.com/) and provides a modern, searchable interface for developers.

## 📁 Structure

```
docs/
├── README.md                 # This file
├── docs.json                 # Mintlify configuration
├── introduction.mdx          # Main introduction page
├── authentication.mdx        # Authentication overview
├── rate-limiting.mdx         # Rate limiting guide
├── errors.mdx               # Error handling guide
├── auth/                     # Authentication endpoints
│   ├── register.mdx
│   ├── login.mdx
│   ├── google.mdx
│   ├── refresh.mdx
│   ├── password-reset.mdx
│   └── invite.mdx
├── analytics/                # Analytics endpoints
│   ├── campaign.mdx
│   ├── email.mdx
│   ├── audience.mdx
│   ├── team.mdx
│   ├── trends.mdx
│   ├── heatmap.mdx
│   └── export.mdx
├── tracking/                 # Email tracking endpoints
│   ├── open.mdx
│   └── click.mdx
├── subscriptions/            # Subscription endpoints
│   ├── create.mdx
│   ├── features.mdx
│   ├── portal.mdx
│   └── webhook.mdx
├── files/                    # File management endpoints
│   └── upload.mdx
├── users/                    # User management endpoints
│   ├── me.mdx
│   ├── list.mdx
│   ├── get.mdx
│   ├── update.mdx
│   └── delete.mdx
├── guides/                   # Tutorial guides
│   ├── quickstart.mdx
│   ├── authentication.mdx
│   ├── analytics.mdx
│   └── webhooks.mdx
├── examples/                 # Code examples
│   ├── nodejs.mdx
│   ├── python.mdx
│   └── curl.mdx
└── changelog/               # Changelog
    └── overview.mdx
```

## 🚀 Getting Started

### Prerequisites
- Node.js 16+ (for local development)
- Mintlify CLI (for local preview)

### Installation

1. **Install Mintlify CLI** (optional, for local development):
   ```bash
   npm install -g mintlify
   ```

2. **Start local development server**:
   ```bash
   mintlify dev
   ```

3. **Build for production**:
   ```bash
   mintlify build
   ```

## 📝 Writing Documentation

### File Naming
- Use kebab-case for file names
- Use `.mdx` extension for all documentation files
- Include frontmatter with title and description

### Frontmatter Format
```mdx
---
title: 'Page Title'
description: 'Brief description of the page content'
---
```

### Code Examples
- Include examples in multiple languages (JavaScript, Python, cURL)
- Use proper syntax highlighting
- Include error handling examples
- Show both success and error responses

### Parameter Documentation
Use the `<ParamField>` component for documenting API parameters:

```mdx
<ParamField query="campaignId" type="string" required>
The unique identifier of the campaign to analyze.
</ParamField>
```

## 🔧 Configuration

The `docs.json` file configures the documentation structure and appearance:

- **Navigation**: Defines the sidebar structure and page organization
- **Styling**: Customizes colors, fonts, and layout
- **Integrations**: Sets up external services and analytics
- **SEO**: Configures meta tags and social sharing

## 📚 Content Guidelines

### API Documentation
- Include complete request/response examples
- Document all parameters and their types
- Show error responses and status codes
- Explain authentication requirements
- Provide rate limiting information

### Guides
- Start with simple examples
- Build up to complex use cases
- Include troubleshooting sections
- Link to related documentation

### Examples
- Use realistic data
- Include error handling
- Show best practices
- Provide complete working code

## 🎨 Styling

### Custom CSS
The `style.css` file contains custom styles for the documentation:

```css
/* Custom styles for code blocks */
.code-block {
  border-radius: 8px;
  margin: 1rem 0;
}

/* Custom button styles */
.cta-button {
  background: linear-gradient(45deg, #26272b, #A1B659);
  color: white;
  padding: 12px 24px;
  border-radius: 6px;
  text-decoration: none;
}
```

### Branding
- Primary color: `#26272b`
- Accent color: `#A1B659`
- Dark mode support
- Custom logo integration

## 🔄 Deployment

### Production Deployment
The documentation is automatically deployed when changes are pushed to the main branch.

### Staging Deployment
Create a pull request to preview changes before merging.

### Custom Domain
The documentation is served at `https://docs.xem.email`

## 📊 Analytics

### Usage Tracking
- Page views and navigation patterns
- Search queries and results
- Time spent on pages
- User feedback and ratings

### Performance Monitoring
- Page load times
- Search performance
- Error rates
- User satisfaction scores

## 🤝 Contributing

### Adding New Endpoints
1. Create a new `.mdx` file in the appropriate directory
2. Follow the existing documentation format
3. Include complete examples and error handling
4. Update the navigation in `docs.json`
5. Test locally before submitting

### Updating Existing Documentation
1. Make changes to the relevant `.mdx` file
2. Ensure all examples still work
3. Update any related links or references
4. Test the changes locally

### Documentation Standards
- Use clear, concise language
- Include practical examples
- Follow the established format
- Test all code examples
- Include error scenarios

## 🆘 Support

### Getting Help
- **Documentation Issues**: Create an issue in the repository
- **Content Questions**: Contact the documentation team
- **Technical Problems**: Check the Mintlify documentation

### Resources
- [Mintlify Documentation](https://mintlify.com/docs)
- [MDX Guide](https://mdxjs.com/)
- [Xem API Reference](https://api.xem.email)

## 📄 License

This documentation is licensed under the MIT License - see the LICENSE file for details.
