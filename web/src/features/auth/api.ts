import { api } from '../../lib/apiClient'
import type { TokenPair, User } from '../../types'

export interface RegisterInput {
  name: string
  email: string
  password: string
}

export interface LoginInput {
  email: string
  password: string
  deviceName?: string
}

export function registerUser(input: RegisterInput): Promise<void> {
  return api.post<void>('/auth/register', input)
}

export function login(input: LoginInput): Promise<TokenPair> {
  return api.post<TokenPair>('/auth/login', {
    email: input.email,
    password: input.password,
    device_name: input.deviceName,
  })
}

export function logout(refreshToken: string): Promise<void> {
  return api.post<void>('/auth/logout', { refresh_token: refreshToken })
}

export function fetchMe(): Promise<User> {
  return api.get<User>('/users/me')
}
