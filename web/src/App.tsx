import { RouterProvider } from 'react-router-dom'
import { ToastViewport } from './components/ui'
import { AuthProvider } from './features/auth/AuthProvider'
import { router } from './routes'

export function App() {
  return (
    <AuthProvider>
      <RouterProvider router={router} />
      <ToastViewport />
    </AuthProvider>
  )
}
