Current State          →        What You'll Add
──────────────────────────────────────────────────
Basic CRUD             →        JWT Authentication
No caching             →        Redis (cache layer)
Single collection      →        Multiple collections + relations
Manual testing         →        Postman collections / unit tests
No middleware          →        Auth middleware, logging, rate limiting
Plain structs          →        Proper error handling & response wrappers



Register Flow
User sends name, email, password
          ↓
Validate input          ← email format, password length
          ↓
Check email exists      ← query MongoDB, return error if duplicate
          ↓
Hash password           ← bcrypt, never store raw password
          ↓
Create user document    ← role = "user", plan = "basic"
          ↓
Save to MongoDB
          ↓
Return success message  ← do NOT return token on register



Login Flow
User sends email, password
          ↓
Find user by email      ← return error if not found
          ↓
Compare password        ← bcrypt compare, return error if wrong
          ↓
Generate Access Token   ← JWT, expires in 15 minutes
          ↓
Generate Refresh Token  ← JWT, expires in 7 days
          ↓
Save Refresh Token      ← store in user document in MongoDB
          ↓
Return both tokens      ← frontend stores these



Refresh Token Flow
Access token expires (15 min)
          ↓
Frontend sends refresh token
          ↓
Validate refresh token signature
          ↓
Find user in MongoDB, compare stored refresh token
          ↓
Generate new access token
          ↓
Return new access token