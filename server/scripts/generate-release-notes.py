import json
import os
import requests

def generate_release_notes():
    try:
        data = {
            "model": "nvidia/nemotron-3.5-lightning:free",
            "messages": [
                {
                    "role": "system", 
                    "content": """You are a helpful assistant that generates user-friendly release notes for Posthoot, an email campaign management platform. Follow these guidelines:

                        1. Write in a clear, conversational tone that non-technical users can understand
                        2. Group changes into categories like '🎉 New Features', '✨ Improvements', '🐛 Bug Fixes', '🔒 Security', and '🔧 Infrastructure'
                        3. Explain changes from the user's perspective - what they can do now that they couldn't before
                        4. Highlight exciting new capabilities in email campaigns, templates, contacts, team management, automation, and analytics
                        5. Use simple, jargon-free language to explain technical changes like authentication, permissions, and API improvements
                        6. Include relevant emojis to make the notes engaging and scannable (📨 for email features, 👥 for team features, 🔐 for auth features, etc.)
                        7. Call out any changes that affect subscription features, SMTP configuration, or domain management
                        8. Add helpful tips for getting the most out of new email automation and campaign features
                        9. Thank users for their feedback and contributions to the email marketing platform
                        10. Keep the tone positive and enthusiastic while being honest about fixes
                        11. Format in an easy-to-read style with clear headings and sections
                        12. Mention integration improvements with Google OAuth, Firebase, or payment systems when relevant
                        13. Highlight features that improve email deliverability, campaign performance, or team collaboration
                        14. Reference specific modules like Campaign Management, Template Management, Contact Management, Team Management, User Management, API Key Management, Automation, SMTP Configuration, Domain Management, or Webhook Management when applicable
                        15. End with what's coming next for the email marketing platform to build excitement

                        Only return the formatted release notes, no other text."""
                },
                {
                    "role": "user",
                    "content": f"Generate detailed release notes with the following information:\n\nCommit History: {os.environ.get('COMMITS')}\nRepository: {os.environ.get('GITHUB_REPOSITORY')}\nRelease Name: {os.environ.get('DOCKERLIKE_RELEASE_NAME')}\nVersion: {os.environ.get('NEW_VERSION')}\n\nPlease analyze the commits and generate comprehensive release notes following the system guidelines."
                }
            ]
        }

        response = requests.post(
            url=os.environ.get("OPENROUTER_BASE_URL") + "/chat/completions",
            headers={
                "Authorization": f"Bearer {os.environ.get('OPENROUTER_API_KEY')}",
                "Content-Type": "application/json",
                "HTTP-Referer": os.environ.get('SITE_URL', ''),
                "X-Title": os.environ.get('SITE_NAME', '')
            },
            data=json.dumps(data)
        )

        response.raise_for_status()
        content = response.json()["choices"][0]["message"]["content"]
        
        with open("release_notes.md", "w") as f:
            f.write(content)

    except Exception as error:
        print(f"Error: {error}")
        exit(1)

if __name__ == "__main__":
    generate_release_notes()
