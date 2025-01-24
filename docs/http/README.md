# API Testing with REST Client

This directory contains HTTP request files for testing the API endpoints using the [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) VS Code extension.

## Setup

1. Install the REST Client extension in VS Code
2. Open any `.http` file
3. Click "Send Request" above any request to execute it

## Environment Variables

The `http-client.env.json` file contains environment-specific variables. To switch environments:

1. Click the "No Environment" dropdown in the bottom right of VS Code
2. Select the desired environment (development, staging, production)

## Available Request Files

### auth.http

Authentication-related endpoints:

- `POST /auth/signup` - Register a new user
- `POST /auth/login` - Login and get tokens
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - Logout and invalidate tokens

### Usage Tips

1. Variables are automatically shared between requests
2. The `@accessToken` and `@refreshToken` are automatically populated after login
3. Use `###` to separate requests
4. Use `@name` to reference responses in other requests

Example workflow:
1. Run signup request (first time only)
2. Run login request
3. Use other endpoints with the automatically populated tokens 