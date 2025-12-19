# GO BITE (gobite)
GoBite is an opinionated, production-ready server API framework written in Go. It is designed to help developers quickly set up robust, scalable, and secure web services with minimal configuration. With a focus on simplicity, extensibility, and best practices, GoBite provides a strong foundation for building and deploying production-grade applications. This is suitable for a wide range of use cases, from microservices to monolithic applications. It comes with a set of pre-configured tools, libraries, and patterns to streamline development while ensuring maintainability, scalability, and performance.

## TODO

### Backend

#### Infrastructure
- [ ] setup storage compatible with S3, GCS, MiniO

#### Auth & Sessions
- [x] login with email/password
- [ ] login with TOTP
- [x] logout
- [x] logout all
- [x] refresh token rotation
- [x] register with email/password
- [ ] register resend verification email
- [ ] login with Google
- [ ] register with Google
- [ ] email verification token payload (subject + purpose)


#### Passwords
- [ ] issue reset token (forgot password)
- [ ] deliver reset email
- [ ] reset password with token
- [ ] verify current password on change

#### MFA (TOTP)
- [ ] setup TOTP
- [ ] confirm TOTP

#### Profile
- [ ] get profile
- [ ] update profile

#### Notifications
- [x] register device token
- [x] remove device token
- [ ] list notifications
- [ ] unread count
- [ ] mark read
- [ ] mark all read
- [ ] delete notification

### Web
