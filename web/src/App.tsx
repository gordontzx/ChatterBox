import { RouterProvider } from 'react-router-dom'
import { ToastViewport } from './components/ui'
import { AuthProvider } from './features/auth/AuthProvider'
import { RealtimeProvider } from './features/realtime/RealtimeProvider'
import { router } from './routes'

export function App() {
  return (
    <AuthProvider>
      <RealtimeProvider>
        <RouterProvider router={router} />
        <ToastViewport />
      </RealtimeProvider>
    </AuthProvider>
  )
}
