/** Mirrors dto.OpeningResponse on the Go side. */
export interface Opening {
  id: number
  createdAt: string
  updatedAt: string
  role: string
  company: string
  location: string
  remote: boolean
  link: string
  salary: number
}

/** Mirrors dto.CreateOpeningRequest / dto.UpdateOpeningRequest. */
export interface OpeningPayload {
  role: string
  company: string
  location: string
  remote: boolean
  link: string
  salary: number
}

/** Mirrors dto.APIResponse — the envelope every endpoint replies with. */
export interface ApiResponse<T> {
  message: string
  data?: T
  error?: string
}
