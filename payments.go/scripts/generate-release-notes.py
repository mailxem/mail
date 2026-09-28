import json
import os
import urllib.request
import urllib.error

def generate_release_notes():
    try:
        data = json.dumps({
            "model": "gpt-4o",
            "messages": [
                {
                    "role": "system", 
                    "content": """You are a helpful assistant that generates user-friendly release notes for a SaaS product. Follow these guidelines:

                        1. Write in a clear, conversational tone that non-technical users can understand
                        2. Group changes into categories like 'New Features', 'Improvements', and 'Bug Fixes'
                        3. Explain changes from the user's perspective - what they can do now that they couldn't before
                        4. Highlight exciting new capabilities and how they benefit users
                        5. Use simple, jargon-free language to explain technical changes
                        6. Include friendly emojis to make the notes engaging and scannable
                        7. Call out any changes that users need to be aware of
                        8. Add helpful tips for getting the most out of new features
                        9. Thank users for their feedback and contributions
                        10. Keep the tone positive and enthusiastic while being honest about fixes
                        11. Format in an easy-to-read style with clear headings
                        12. Include visuals like screenshots where helpful
                        13. Add links to documentation for those who want more details
                        14. Highlight customer-requested features that were implemented
                        15. End with what's coming next to build excitement

                        Only return the formatted release notes, no other text."""
                },
                {
                    "role": "user",
                    "content": f"Generate detailed release notes with the following information:\n\nCommit History: {os.environ.get('COMMITS')}\nRepository: {os.environ.get('GITHUB_REPOSITORY')}\nRelease Name: {os.environ.get('DOCKERLIKE_RELEASE_NAME')}\nVersion: {os.environ.get('NEW_VERSION')}\n\nPlease analyze the commits and generate comprehensive release notes following the system guidelines."
                }
            ]
        })

        url = os.environ.get("OPENAI_BASE_URL") + "/chat/completions"
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {os.environ.get('OPENAI_API_KEY')}"
        }

        req = urllib.request.Request(
            url,
            data=data.encode("utf-8"),
            headers=headers,
            method="POST"
        )

        with urllib.request.urlopen(req) as response:
            response_data = response.read()
            content = json.loads(response_data)["choices"][0]["message"]["content"]
            with open("release_notes.md", "w") as f:
                f.write(content)

    except Exception as error:
        print(f"Error: {error}")
        exit(1)

if __name__ == "__main__":
    generate_release_notes()
