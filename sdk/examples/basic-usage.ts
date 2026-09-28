import { Xem } from '../src';

async function main() {
    // Initialize the SDK
    const xem = new Xem();

    // Login with API key
    xem.login('your-api-key-here');

    // Or initialize with API key directly:
    // const xem = new Xem({ apiKey: 'your-api-key-here' });

    try {
        // Send a basic email
        const result = await xem.email.send({
            to: 'recipient@example.com',
            subject: 'Hello from XEM SDK!',
            html: '<h1>Welcome</h1><p>This email was sent using the XEM SDK.</p>',
        });

        console.log('Email sent successfully:', result);
    } catch (error: any) {
        console.error('Failed to send email:', error.message);
    }
}

main();
