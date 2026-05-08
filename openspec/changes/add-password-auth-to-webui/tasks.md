## 1. Web Profile Inputs

- [x] 1.1 Add an SSH password field to the WebUI profile form, request parsing, and profile data model.
- [x] 1.2 Ensure saved profiles can persist and reload the SSH password value without breaking existing profile files.

## 2. Authentication Priority

- [x] 2.1 Update runtime configuration mapping so Web-managed tunnels pass the SSH password through to `TunnelConfig`.
- [x] 2.2 Change SSH auth method selection so password authentication is used first when present, and private key authentication is only used when password is empty.

## 3. Verification

- [x] 3.1 Add or update tests for Web profile save/load behavior with passwords.
- [x] 3.2 Add or update tests for password-priority SSH authentication behavior and fallback to private key when password is absent.
