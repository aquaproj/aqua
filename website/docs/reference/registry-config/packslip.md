---
sidebar_position: 2115
---

# packslip

- `aqua >= v2.65.0`

Please see [Reference](/docs/reference/security/packslip) too.

## Fields

- enabled (bool)
- pubkey (string): The public key of a key-signed release, as its base64 line. `packslip verify --pubkey`
- identity (string): The exact certificate identity a keyless signer must have, including its ref. `packslip verify --identity`
- identity_prefix (string): A prefix the certificate identity must start with, such as `https://github.com/owner/repo/`. `packslip verify --identity-prefix`
- issuer (string): The OIDC issuer a keyless signer must have. `packslip verify --issuer`
- pin (string): The signer fingerprint (`ps1_...`) the certificate's repository must have. `packslip verify --pin`
- allow_unlogged (bool): Accept a bundle without a transparency log entry. `packslip verify --allow-unlogged`

At least one of `pubkey`, `identity`, `identity_prefix`, and `pin` is required.

e.g.

```yaml
packslip:
  identity_prefix: https://github.com/jdx/hk/
  issuer: https://token.actions.githubusercontent.com
```
