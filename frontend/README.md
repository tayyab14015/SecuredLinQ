# SecuredLinQ Frontend

React SPA frontend for the SecuredLinQ video call and load management application.

## Technology Stack

- **React 18** with TypeScript
- **Vite** for fast development and building
- **React Router** for client-side routing
- **Tailwind CSS** for styling
- **Agora RTC SDK** for video calls
- **Axios** for API requests

## Features

- Admin dashboard for managing loads
- Video call interface with:
  - Two-way video/audio communication
  - Camera switching (for drivers) when on mobile
  - Screenshot capture (admin only)
  - Cloud recording (admin only)
- Cookie-based session authentication

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

```bash
# Install dependencies
npm install


# Update .env with your configuration
```

### Development

```bash
# Start development server
npm run dev
```

The app will be available at `http://localhost:5173`.

### Building

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

## Project Structure

```
frontend/
├── src/
│   ├── api/              # API client and types
│   ├── components/       # Reusable components
│   ├── context/          # React context providers
│   ├── pages/            # Page components
│   │   ├── admin/        # Admin pages
│   │   └── ...           # Driver pages
│   ├── App.tsx           # Main app with routing
│   ├── main.tsx          # Entry point
│   └── index.css         # Global styles
├── index.html
├── vite.config.ts
└── tailwind.config.js
```

## Pages

### Admin Pages

- `/admin/login` - Admin login page
- `/admin/dashboard` - Dashboard with load list
- `/admin/load/:guestRand` - Load detail with media gallery
- `/admin/meeting/:roomId` - Admin video call interface

### Driver Pages

- `/` - Entry point (parses JWT token from URL)
- `/Load/:guestRand` - Join meeting page
- `/join/:roomId` - Video call interface

## Environment Variables

| Variable | Description |
|----------|-------------|
| `VITE_API_URL` | Backend API URL (leave empty for same-origin) |
| `VITE_AGORA_APP_ID` | Agora App ID for video calls |

## API Integration

The frontend communicates with the Go backend via REST APIs. Cookie-based authentication is used for admin sessions.

### Authentication Flow

1. Admin logs in via `/api/auth/login`
2. Server sets `session_id` cookie (HttpOnly, Secure, SameSite)
3. All subsequent requests include the cookie automatically
4. Protected routes check session validity

### Video Call Flow

1. Admin creates meeting room for a load
2. Meeting link is sent to driver via SMS
3. Both parties join the Agora channel using tokens from backend
4. Admin can capture screenshots and record the call
5. Recordings are saved to S3

## Styling

The project uses Tailwind CSS with custom CSS variables for theming:

```css
:root {
  --primary: #1e3a5f;
  --accent: #f59e0b;
  --success: #10b981;
  --error: #ef4444;
  /* ... */
}
```

Custom utility classes are defined in `index.css`:
- `.btn`, `.btn-primary`, `.btn-accent`, etc. for buttons
- `.card` for card containers
- `.input` for form inputs
- `.table` for data tables
