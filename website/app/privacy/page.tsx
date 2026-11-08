import Link from "next/link";
import { PrivacyLayout } from "@/components/privacy-layout";
import {
  privacyContact,
  privacyEmailLink,
  privacyMetadata,
} from "@/lib/privacy";

export const metadata = privacyMetadata(
  "Privacy policy — draft",
  "How Xem handles account, workspace and connected email data, with proposed retention commitments and your privacy choices.",
  "/privacy",
);

export default function PrivacyPage() {
  return (
    <PrivacyLayout
      eyebrow="Your data, explained"
      title="Email is personal. So is privacy."
      description="What Xem needs to work, where information goes, and the choices you keep. Written to be read."
      summary={
        <>
          <p className="font-medium">The proposed commitments</p>
          <ul className="mt-4 space-y-3">
            <li>Use your data to provide the features you choose.</li>
            <li>
              Do not sell Google user data, use it for advertising, or use it to
              train general-purpose AI models.
            </li>
            <li>
              Remove disconnected mailbox credentials from the active database.
              Treat deletion of saved messages as a separate request.
            </li>
            <li>
              Complete verified deletion requests within 30 days, with remaining
              backup copies expiring within 90 days after active-system
              deletion, subject to the limited exceptions below.
            </li>
          </ul>
        </>
      }
      sections={[
        {
          id: "who-we-are",
          title: "Who this covers",
          content: (
            <>
              <p>
                This policy is for Xem’s website and hosted service, operated
                under the Xem name. Contact the Xem team at{" "}
                <a href={privacyEmailLink}>{privacyContact}</a> about privacy,
                access, corrections, or deletion. The operator’s full legal
                identity will be added before this draft becomes effective.
              </p>
              <p>
                When an organization uses Xem, it controls the contacts,
                campaigns and workspace content it adds. Direct requests about
                an organization’s mailing list to that organization; we can help
                identify the right route. A separately self-hosted installation
                is controlled by its operator, who must provide their own
                privacy notice and retention policy. Merely running the
                open-source software does not give Xem access to that
                installation’s data.
              </p>
            </>
          ),
        },
        {
          id: "data",
          title: "What we collect and why",
          content: (
            <>
              <ul>
                <li>
                  <strong>Account and workspace details:</strong> name, email
                  address, membership, permissions and authentication records
                  needed to sign in and control access.
                </li>
                <li>
                  <strong>Content you provide:</strong> contacts, lists, forms,
                  templates, campaign content, outgoing messages and
                  attachments, so we can store, display and send them as
                  directed.
                </li>
                <li>
                  <strong>Sending and engagement records:</strong> recipients,
                  delivery status, errors, timestamps and, when enabled by the
                  sender, tracked opens and clicks. These help senders manage
                  delivery and reporting; an open is not proof a person read an
                  email.
                </li>
                <li>
                  <strong>Service and security information:</strong> request
                  information such as IP address, browser information and
                  diagnostic events needed to operate the service, investigate
                  faults and prevent abuse.
                </li>
                <li>
                  <strong>Support correspondence:</strong> information you
                  choose to share when asking for help. Do not send passwords or
                  API tokens.
                </li>
              </ul>
              <p>
                Sign-in uses session information. Providers that host or deliver
                the website may process connection information. Our marketing
                website uses Google Analytics (the Google tag) to count visits
                and see which pages are used; it sets cookies and sends Google
                your IP address, browser details and the pages you view. We
                don’t use it for advertising, and it isn’t loaded inside the
                Xem app or on newsletter sites. This draft does not claim a
                completed production-cookie audit.
              </p>
            </>
          ),
        },
        {
          id: "google",
          title: "Connected Google mail",
          content: (
            <>
              <p>
                Ordinary Google sign-in and permission to access Gmail are
                separate choices. The Google mail feature is in limited testing.
                Connecting a mailbox authorizes Xem’s backend to access the
                selected Gmail or Google Workspace account to display folders,
                search and read messages and attachments, change read and star
                flags, and send messages and replies you request. The connected
                sender can also be used in workflows you configure, subject to
                Google’s limits and anti-spam requirements.
              </p>
              <p>
                The current test connector uses Google’s full-mailbox permission
                for IMAP and SMTP. This permission is broader than the
                operations currently offered. It is under review before public
                availability. It does not provide access to Drive files,
                calendars, or all accounts in your Google Workspace domain.
              </p>
              <p>
                Incoming mail is fetched on demand; the connector does not
                maintain a complete local mirror of the inbox. The backend
                stores the mailbox address, configuration references and
                encrypted authorization credentials. Outgoing message records
                can contain subject, body, recipients, attachments and delivery
                information. Opening mail also sends the requested content to
                your browser.
              </p>
              <p>
                <strong>
                  A connected mailbox is shared with the Xem workspace under its
                  access permissions.
                </strong>{" "}
                It is not a private mailbox visible only to the person who
                connects it. Connect only an account you are authorized to
                share. Learn more about{" "}
                <Link href="/google-mail">
                  Google mail access and disconnecting
                </Link>
                .
              </p>
            </>
          ),
        },
        {
          id: "sharing",
          title: "Who can receive data",
          content: (
            <>
              <p>
                Data may be shared with authorized workspace members; email
                providers and intended recipients when you send messages;
                infrastructure providers that host, store and secure the
                service; and integrations you deliberately enable, including
                webhooks, MCP clients and optional AI features. What each
                integration receives depends on its permissions and the action
                you request.
              </p>
              <p>
                Our proposed commitments require service providers to use data
                only to supply the requested service under appropriate
                restrictions. We do not sell Google user data or use it for
                advertising, data brokerage, credit decisions, or unrelated
                profiling. We may disclose data when legally required or
                necessary to investigate abuse or a security incident, limited
                to that purpose. Any transfer of Google data as part of a
                business sale requires prior user consent where Google’s policy
                requires it.
              </p>
              <p>
                Staff access to Google user data is limited to your documented
                consent to specific support access, necessary security work,
                legally required access, or permitted anonymized and aggregated
                internal operations. Support access is not permission to browse
                your mailbox.
              </p>
            </>
          ),
        },
        {
          id: "ai",
          title: "Optional AI features",
          content: (
            <>
              <p>
                Using the writing assistant sends your instructions and current
                draft, including its subject and body, to the configured AI
                service. Chat sends your messages and relevant results returned
                by workspace tools. Those results can include outgoing email
                data. Google-derived text may be included if you ask AI to work
                with a reply or other email content. Connecting Gmail by itself
                does not request an AI summary of the inbox.
              </p>
              <p>
                Our proposed policy prohibits using Google user data to train or
                improve general-purpose AI models. Hosted AI providers, their
                retention periods and their contractual data-use terms must be
                confirmed and disclosed before this policy becomes effective or
                Google data is enabled for hosted AI processing. Those checks
                are not complete; the name of a model is not evidence of its
                provider’s data handling.
              </p>
              <p>
                You can read, compose and send messages without using AI. Review
                the disclosure beside an AI action before using it. A
                third-party MCP or AI client you connect has its own privacy
                practices; choose it carefully.
              </p>
            </>
          ),
        },
        {
          id: "retention",
          title: "Retention and deletion",
          content: (
            <>
              <p>
                Account and workspace content stays available while needed to
                provide the service, until you delete it or make a verified
                deletion request. Hiding or soft-deleting a record is not
                necessarily physical erasure. Disconnecting Google removes the
                saved authorization secret from the active database and disables
                the linked sender and inbox; it does not automatically erase
                outbox messages, the connection’s address, other workspace
                records, or backups.
              </p>
              <ul>
                <li>
                  <strong>Verified deletion requests:</strong> proposed
                  completion within 30 calendar days of verifying your identity
                  and the scope of the request.
                </li>
                <li>
                  <strong>Backups:</strong> proposed expiry of residual copies
                  within 90 calendar days after deletion from active systems.
                  Backup copies are isolated from normal product use, and
                  completed deletions must be reapplied before restored data
                  returns to service.
                </li>
                <li>
                  <strong>Assistant history:</strong> the application expires
                  saved conversations seven days after the last save. Deleting a
                  conversation removes it from active history. Provider-side
                  records and backups require separate retention controls.
                </li>
                <li>
                  <strong>Limited exceptions:</strong> records required by law,
                  a documented security investigation or a legal claim may be
                  retained for that specific purpose. We will explain the
                  category and applicable period when permitted. Exceptions do
                  not justify keeping an entire mailbox indefinitely.
                </li>
              </ul>
              <p>
                These are proposed service commitments, not claims that every
                production datastore or backup has already been configured. The
                operational checks must pass before this draft takes effect.{" "}
                <Link href="/data-deletion">See how to request deletion</Link>.
              </p>
            </>
          ),
        },
        {
          id: "security",
          title: "Protecting your information",
          content: (
            <>
              <p>
                The connected-mail implementation encrypts provider credentials
                on the backend and uses verified secure connections to Google.
                Workspace permissions control access, and only administrators
                can manage mailbox connections. The application does not return
                saved provider secrets in connection settings.
              </p>
              <p>
                These controls do not make all message content end-to-end
                encrypted or eliminate security risks. Hosted infrastructure,
                access reviews, backup controls and incident procedures must
                also be validated. We do not claim Google certification, a
                completed independent security assessment or a security
                certification that has not been earned.
              </p>
            </>
          ),
        },
        {
          id: "choices",
          title: "Your choices and rights",
          content: (
            <>
              <p>
                You may ask to access, correct, export or delete your personal
                information, and raise concerns about how it is used. Depending
                on where you live, you may also have rights to restrict or
                object to processing and to complain to a privacy regulator.
                Contact <a href={privacyEmailLink}>{privacyContact}</a>; we
                verify requests using information proportionate to the request.
              </p>
              <p>
                You can disconnect a Google mailbox in Xem and revoke the grant
                in your{" "}
                <a href="https://myaccount.google.com/connections">
                  Google Account connections
                </a>
                . Revoking access does not delete messages already stored by Xem
                or recipients. Deletion requests for a shared workspace require
                checking the requester’s authority and other members’ rights.
              </p>
            </>
          ),
        },
        {
          id: "limited-use",
          title: "Google Limited Use",
          content: (
            <>
              <p>
                The use of information received from Google Workspace scopes
                will adhere to the{" "}
                <a href="https://developers.google.com/terms/api-services-user-data-policy">
                  Google API Services User Data Policy
                </a>
                , including the Limited Use requirements, and the{" "}
                <a href="https://developers.google.com/workspace/workspace-api-user-data-developer-policy">
                  Google Workspace API User Data and Developer Policy
                </a>
                .
              </p>
              <p>
                This is a proposed commitment to those restrictions. It is not a
                statement that Google has approved Xem. Public hosted Google
                access remains unavailable until the applicable verification and
                security assessment are complete.
              </p>
            </>
          ),
        },
        {
          id: "changes",
          title: "Changes to this policy",
          content: (
            <p>
              We will publish an effective date when this policy is ready.
              Material changes will be explained before taking effect, and we
              will obtain additional consent where required for a new use of
              Google data. The review date above identifies this draft only.
            </p>
          ),
        },
      ]}
    />
  );
}
