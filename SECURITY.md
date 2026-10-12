# Security Policy

If you discover a security vulnerability, please do not open a public issue. Instead, report it privately via the process outlined below to ensure vulnerabilities are patched before they are disclosed.

## Prohibited Content

Please do not transmit personal data, secrets, nonpublic information, malicious code, or any data that could create regulatory compliance obligations, legal liabilities, or other risks for the recipient.

## Vulnerability Reporting Process

Please note that unvetted AI-generated security reports and reproducers can delay fixes for real bugs.

1. Before reporting a bug, please ensure a human confirms it matters in real use and any reproducer code is valid.  Some AI-detected bugs only occur when AI-written reproducers use anti-patterns or settings that don't reflect real user apps in production environments.

2. Email a brief report and minimal reproducer to faye.github@gmail.com. To reduce delays, follow these guidelines:
   - Begin the email subject with "fxamacker/cbor vulnerability".
   - Keep the report accurate and preferably under one page (omit AI-generated filler).
   - Include the results of testing the reproducer with both the latest release and the current main branch.
   - Write the reproducer as a single file that doesn't import other third-party packages.
   - Use the simplest data types and data structures that reproduce the bug.
   - Keep the CBOR data as small as possible (max 4096 bytes of CBOR data).
   - Do not send download links or binary files (attach only .go or .txt files).

3. Do not disclose the vulnerability publicly until a fix is released or until the disclosure date agreed with the maintainer, whichever comes first.  Even a one-line security fix requires testing and fuzzing before the pull request is opened in public.

## Support Agreements with Response Times

If you require a guaranteed response time or rely on this library for mission-critical operations, please consider a dedicated support agreement or sponsorship to fund ongoing maintenance and improvements. Client-specific features can be accommodated in separate branches or private repositories.
