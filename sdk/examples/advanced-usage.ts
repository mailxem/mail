import { Xem } from '../src';

async function main() {
    const xem = new Xem({ apiKey: 'your-api-key-here' });

    // Example 1: Email with template
    try {
        await xem.email.send({
            to: 'user@example.com',
            subject: 'Welcome to our platform!',
            templateId: 'welcome-template',
            data: [
                { name: 'John Doe', signupDate: '2026-02-16' }
            ],
        });
        console.log('Template email sent');
    } catch (error: any) {
        console.error('Template email failed:', error.message);
    }

    // Example 2: Scheduled email
    try {
        await xem.email.send({
            to: 'user@example.com',
            subject: 'Upcoming event reminder',
            html: '<p>Don\'t forget about the event tomorrow!</p>',
            scheduleAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(), // Tomorrow
        });
        console.log('Scheduled email queued');
    } catch (error: any) {
        console.error('Scheduled email failed:', error.message);
    }

    // Example 3: Email with CC, BCC, and Reply-To
    try {
        await xem.email.send({
            to: 'primary@example.com',
            cc: 'manager@example.com',
            bcc: 'archive@example.com',
            replyTo: 'support@example.com',
            subject: 'Monthly Report',
            html: '<h2>Monthly Report</h2><p>Please find the report attached.</p>',
        });
        console.log('Email with CC/BCC sent');
    } catch (error: any) {
        console.error('CC/BCC email failed:', error.message);
    }

    // Example 4: Test mode (email won't be sent)
    try {
        await xem.email.send({
            to: 'test@example.com',
            subject: 'Test Email',
            html: '<p>This is a test</p>',
            test: true,
        });
        console.log('Test email validated (not sent)');
    } catch (error: any) {
        console.error('Test email failed:', error.message);
    }

    // Example 5: Custom provider
    try {
        await xem.email.send({
            to: 'user@example.com',
            subject: 'Custom Provider Email',
            html: '<p>Sent via custom provider</p>',
            provider: 'CUSTOM',
        });
        console.log('Custom provider email sent');
    } catch (error: any) {
        console.error('Custom provider email failed:', error.message);
    }
}

main();
