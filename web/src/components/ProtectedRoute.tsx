import type { ReactNode } from 'react'

// TODO(M1): redirect to /login when the AuthProvider says we're unauthenticated.
// Pass-through for M0 so routing/layout can be verified before auth exists.
export function ProtectedRoute({ children }: { children: ReactNode }) {
  return <>{children}</>
}
