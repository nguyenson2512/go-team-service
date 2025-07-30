# Team Service - Clean Architecture

This project has been refactored to follow Clean Architecture principles, providing better separation of concerns, maintainability, and testability.

## Architecture Layers

### 1. Entities (`internal/entities/`)
- Core business entities and domain models
- Independent of external concerns
- Contains business rules and logic

### 2. Use Cases (`internal/usecases/`)
- Application business rules
- Orchestrates data flow between entities and repositories
- Contains application-specific business logic

### 3. Repository (`internal/repository/`)
- Data access layer
- Abstracts database operations
- Implements interfaces defined by use cases

### 4. Delivery (`internal/delivery/http/`)
- HTTP handlers and routing
- Converts HTTP requests to use case calls
- Handles HTTP-specific concerns

### 5. Infrastructure (`pkg/`)
- External concerns like database, logging, middleware
- Shared utilities across the application

## API Endpoints

All endpoints remain the same as before:

### Asset Management
- `POST /folders` - Create folder
- `GET /folders/:folderId` - Get folder
- `PUT /folders/:folderId` - Update folder
- `DELETE /folders/:folderId` - Delete folder
- `POST /folders/:folderId/notes` - Create note
- `GET /notes/:noteId` - Get note
- `PUT /notes/:noteId` - Update note
- `DELETE /notes/:noteId` - Delete note

### Sharing
- `POST /folders/:folderId/share` - Share folder
- `DELETE /folders/:folderId/share/:userId` - Revoke folder share
- `POST /notes/:noteId/share` - Share note
- `DELETE /notes/:noteId/share/:userId` - Revoke note share

### Team Management
- `POST /teams` - Create team
- `POST /teams/:teamId/members` - Add member
- `DELETE /teams/:teamId/members/:memberId` - Remove member
- `POST /teams/:teamId/managers` - Add manager
- `DELETE /teams/:teamId/managers/:managerId` - Remove manager

### Manager APIs
- `GET /teams/:teamId/assets` - Get team assets
- `GET /users/:userId/assets` - Get user assets

## Environment Variables

Create a `.env` file with the following variables:

```env
DATABASE_DSN=postgres://user:password@localhost:5432/teamservice?sslmode=disable
ACCESS_TOKEN_SECRET=your-secret-key-here
PORT=:8080
```

## Running the Application

1. **Install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Build the application:**
   ```bash
   go build -o bin/app ./cmd/app
   ```

3. **Run the application:**
   ```bash
   ./bin/app
   ```

   Or run directly:
   ```bash
   go run ./cmd/app
   ```

## Benefits of Clean Architecture

1. **Independence**: Business logic is independent of frameworks, UI, and databases
2. **Testability**: Easy to test business logic without external dependencies
3. **Maintainability**: Clear separation of concerns makes code easier to maintain
4. **Flexibility**: Easy to swap out implementations (e.g., change database)
5. **Scalability**: Well-organized code structure supports team growth