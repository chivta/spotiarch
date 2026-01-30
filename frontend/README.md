# Spotiarch Frontend

A Vue.js 3 frontend application with TypeScript, Vue Router, and Pinia for state management.

## Features

- Landing page with call-to-action
- Authentication (Login/Signup)
- Protected dashboard with statistics and items
- Route guards for authentication
- Centralized state management with Pinia
- TypeScript for type safety
- Simple, clean design

## Project Structure

```
src/
├── assets/          # Global styles
├── router/          # Vue Router configuration
├── stores/          # Pinia stores (state management)
├── services/        # API service layer
├── types/           # TypeScript type definitions
├── views/           # Page components
│   ├── LandingPage.vue
│   ├── AuthPage.vue
│   └── Dashboard.vue
├── App.vue          # Root component
└── main.ts          # Application entry point
```

## Getting Started

### Prerequisites

- Node.js (v18 or higher)
- npm or yarn

### Installation

1. Install dependencies:
```bash
npm install
```

2. Create environment file:
```bash
cp .env.example .env
```

3. Update `.env` with your API base URL:
```
VITE_API_BASE_URL=http://localhost:3000/api
```

### Development

Run the development server:
```bash
npm run dev
```

The application will be available at `http://localhost:5173`

### Build

Build for production:
```bash
npm run build
```

Preview production build:
```bash
npm run preview
```

## API Integration

The frontend expects a REST API backend. See [API_DOCUMENTATION.md](./API_DOCUMENTATION.md) for the complete API specification.

Key endpoints:
- `POST /api/auth/signup` - User registration
- `POST /api/auth/login` - User authentication
- `POST /api/auth/logout` - User logout
- `GET /api/auth/me` - Get current user
- `GET /api/dashboard/stats` - Get dashboard statistics
- `GET /api/dashboard/items` - Get dashboard items

## State Management

The app uses Pinia for state management:

- **Auth Store** (`src/stores/auth.ts`): Manages authentication state, user data, and auth-related actions

## Routing

Routes are configured in `src/router/index.ts`:

- `/` - Landing page (public)
- `/auth` - Login/Signup page (public, redirects to dashboard if authenticated)
- `/dashboard` - User dashboard (protected, requires authentication)

## Authentication Flow

1. User enters credentials on `/auth` page
2. Frontend sends request with `credentials: 'include'` to `/api/auth/login` or `/api/auth/signup`
3. Backend sets JWT token as HttpOnly cookie and returns user data
4. Browser automatically stores the HttpOnly cookie
5. All subsequent requests include the cookie automatically (via `credentials: 'include'`)
6. Protected routes check authentication status via router guards
7. Logout calls `/api/auth/logout` which clears the cookie, then redirects to landing page
8. **Automatic 401 handling**: If backend returns 401 (Unauthorized), the frontend automatically:
   - Redirects user to landing page
   - Cookie is cleared by the browser/backend
   - This handles session expiration gracefully

## Technologies

- **Vue 3** - Progressive JavaScript framework
- **TypeScript** - Type safety
- **Vite** - Fast build tool and dev server
- **Vue Router** - Official routing library
- **Pinia** - Official state management library
- **Fetch API** - HTTP requests

## Best Practices Implemented

- Component composition API (script setup)
- TypeScript for type safety
- Centralized API service layer with automatic 401 handling
- Route guards for authentication
- **Secure authentication with HttpOnly cookies** (prevents XSS attacks)
- `credentials: 'include'` for cross-origin cookie transmission
- Environment variables for configuration
- Clean separation of concerns (views, stores, services)
- Responsive design
- Error handling and loading states
- Automatic session expiration handling
