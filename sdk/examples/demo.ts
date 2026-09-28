#!/usr/bin/env node

/**
 * Demo script showing the exact usage pattern requested:
 * const xem = new Xem();
 * xem.login(apiKey);
 * xem.email.send();
 */

import { Xem } from '../src';

async function demo() {
    console.log('🚀 XEM SDK Demo\n');

    // Step 1: Create a new Xem instance
    console.log('Step 1: Initialize SDK');
    const xem = new Xem();
    console.log('✓ SDK initialized\n');

    // Step 2: Login with API key
    console.log('Step 2: Login with API key');
    const apiKey = process.env.XEM_API_KEY || 'your-api-key-here';
    xem.login(apiKey);
    console.log('✓ Logged in\n');

    // Check authentication status
    console.log('Authentication status:', xem.isAuthenticated() ? '✓ Authenticated' : '✗ Not authenticated');
    console.log();

    // Step 3: Send an email
    console.log('Step 3: Send email');
    try {
        const result = await xem.email.send({
            to: 'recipient@example.com',
            subject: 'Hello from XEM SDK!',
            html: `
        <!DOCTYPE html>
        <html>
          <head>
            <style>
              body { font-family: Arial, sans-serif; }
              .container { max-width: 600px; margin: 0 auto; padding: 20px; }
              .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px; }
              .content { padding: 30px 0; }
              .footer { color: #666; font-size: 12px; padding-top: 20px; border-top: 1px solid #eee; }
            </style>
          </head>
          <body>
            <div class="container">
              <div class="header">
                <h1>Welcome to XEM!</h1>
              </div>
              <div class="content">
                <h2>Email Successfully Sent</h2>
                <p>This email was sent using the XEM SDK with the following pattern:</p>
                <pre><code>const xem = new Xem();
xem.login(apiKey);
xem.email.send({...});</code></pre>
                <p>The SDK is working perfectly! 🎉</p>
              </div>
              <div class="footer">
                <p>Sent via XEM Email API</p>
                <p>Timestamp: ${new Date().toISOString()}</p>
              </div>
            </div>
          </body>
        </html>
      `,
            test: true, // Set to false to actually send
        });

        console.log('✓ Email sent successfully!');
        console.log('Response:', result);
    } catch (error: any) {
        console.error('✗ Failed to send email:', error.message);
        if (error.statusCode) {
            console.error('Status code:', error.statusCode);
        }
        if (error.details) {
            console.error('Details:', error.details);
        }
    }

    console.log('\n✨ Demo complete!');
}

// Run the demo
demo().catch(console.error);
