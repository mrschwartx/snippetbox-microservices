export interface User {
  id: number
  name: string
  email: string
  created: string
}

export interface SignupInput {
  name: string
  email: string
  password: string
}

export interface LoginInput {
  email: string
  password: string
}
