# raw-twchat

This is a raw, but still convenient CLI client for Twitch chat. The only two automated inputs are the password, which is taken from environment, and PONGs so you don't have to reply to the PINGs manually. All is shown, except the password, which is obfuscated. Allows replacing unprintable characters with an escaped representation to avoid breaking formatting and invisible characters sneaking through.

## Usage

Example: `TW_OAUTH=xyzxyzxyzxyzxyz raw-twchat-cli -replace-unprintables`

### Options

```
  -insecure
        use a plaintext connection
  -replace-unprintables
        replaces unprintable characters like SOH with \uXXXX representation
```
