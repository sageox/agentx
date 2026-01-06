# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| 0.2.x   | :white_check_mark: |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it by emailing security@sageox.com.

Please do not open a public GitHub issue for security vulnerabilities.

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 7 days
- **Resolution timeline**: Provided after assessment

We will work with you to understand and address the issue. Once resolved, we will coordinate disclosure timing with you.

## Security Considerations

This library detects and interacts with AI coding agents. When using agentx:

- **Environment variables**: The library reads environment variables for agent detection. Ensure your environment is trusted.
- **File system access**: Configuration and context file paths are derived from user input and environment. Validate paths in security-sensitive applications.
- **Hook execution**: When using hook managers, be aware that hooks execute code. Only install hooks from trusted sources.

## Best Practices

When using agentx in your applications:

1. Validate any paths before passing to the library
2. Run with minimal necessary permissions
3. Keep the library updated to the latest version
4. Review hook and command content before installation
