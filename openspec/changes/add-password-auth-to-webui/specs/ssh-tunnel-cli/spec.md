## MODIFIED Requirements

### Requirement: Fatal handling for authentication failures
The tool SHALL treat authentication failure as fatal and MUST exit without entering the reconnect loop. When both an SSH password and a private key path are configured, the tool SHALL prioritize password authentication; it SHALL only use private key authentication when the password is empty.

#### Scenario: Stop on invalid credentials
- **WHEN** the SSH server rejects the supplied credentials during authentication
- **THEN** the tool reports an authentication failure and terminates instead of retrying

#### Scenario: Prefer password when both authentication inputs are configured
- **WHEN** the tool is started with both a non-empty SSH password and a private key path
- **THEN** the tool attempts SSH authentication using the password-based method instead of falling back to the private key first

#### Scenario: Use private key when password is absent
- **WHEN** the tool is started without an SSH password and with a private key path
- **THEN** the tool attempts SSH authentication using the private key
