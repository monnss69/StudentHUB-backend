# StudentHub - Backend

A robust Go-based backend service for the StudentHub platform, built with Gin framework and PostgreSQL.

## 🚀 Features

- User authentication with JWT
- Image upload with Cloudinary integration
- PostgreSQL database integration with GORM
- RESTful API endpoints
- Category and tag management
- Post and comment functionality
- Secure password hashing
- CORS configuration

## 🛠️ Prerequisites

Before you begin, ensure you have installed:
- Go (version 1.23 or higher)
- Git

## 📦 Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/monnss69/StudentHUB-backend
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Create a `.env` file in the root directory:
   ```env
   # Database Configuration
   SUPABASE_DATABASE_URL=postgres://username:password@host:port/database

   # JWT Configuration
   JWT_SECRET=your_jwt_secret_key

   # Cloudinary Configuration
   CLOUD_NAME=your_cloudinary_cloud_name
   CLOUD_API_KEY=your_cloudinary_api_key
   CLOUD_API_SECRET=your_cloudinary_api_secret

   For CVWO reviewer, this is specified in my final write-up letter
   ```

## 🔧 Development

To start the development server:

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## 📚 API Endpoints

### Authentication
- `POST /login` - User login
- `POST /logout` - User logout
- `POST /users` - Create new user

### Users
- `GET /users` - List all users
- `GET /users/:id` - Get user by ID
- `PUT /users/:id` - Update user
- `DELETE /users/:id` - Delete user

### Posts
- `POST /posts` - Create new post
- `GET /posts/:id` - Get post by ID
- `GET /posts/category/:category/:pageIndex` - List posts by category
- `PUT /posts/:id` - Update post
- `DELETE /posts/:id` - Delete post

### Comments
- `GET /posts/:id/comments` - Get post comments
- `POST /posts/:id/comments` - Create comment

### Tags
- `GET /tags` - List all tags
- `POST /posts/:id/tags` - Add tags to post
- `DELETE /posts/:id/tags/:tag_id` - Remove tag from post

### Categories
- `GET /categories` - List all categories
- `GET /categories/:id` - Get category by ID

### Image Upload
- `POST /api/cloudinary/upload` - Upload image
- `DELETE /api/cloudinary/upload/:username` - Delete user image

## 🏗️ Database Schema

The application uses PostgreSQL with the following main tables:
- users
- posts
- comments
- categories
- tags
- posts_tags (junction table)

## 📝 Development Notes

- The server uses Gin framework for routing and middleware
- CORS is configured to allow requests from specified origins
- File uploads are limited to images and have a size limit of 5MB
- JWT tokens are valid for 30 days
- The application includes request retry mechanisms with exponential backoff
