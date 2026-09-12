import { createBrowserRouter } from 'react-router-dom'
import { ProtectedRoute } from './components/ProtectedRoute'
import { LoginPage } from './features/auth/LoginPage'
import { RegisterPage } from './features/auth/RegisterPage'
import { ChatsLayout } from './features/chats/ChatsLayout'
import { EmptyChatsPage } from './features/chats/EmptyChatsPage'
import { ConversationPage } from './features/messages/ConversationPage'

export const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <ChatsLayout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <EmptyChatsPage /> },
      { path: 'chats/:chatId', element: <ConversationPage /> },
    ],
  },
])
