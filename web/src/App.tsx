import { RouterProvider } from 'react-router-dom'
import { ToastViewport } from './components/ui'
import { router } from './routes'

export function App() {
  return (
    <>
      <RouterProvider router={router} />
      <ToastViewport />
    </>
  )
}
