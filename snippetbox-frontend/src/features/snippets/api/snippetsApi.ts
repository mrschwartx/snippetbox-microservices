import { apiClient } from '@/lib/apiClient'
import { NoRecordError, UnauthenticatedError } from '@/lib/errors'
import { sessionStore } from '@/lib/sessionStore'
import { validateSnippetDraft } from '../lib/validation'
import type { Snippet, SnippetDraft } from '../types'

export const getLatestSnippets = async (): Promise<Snippet[]> =>
  apiClient.get<Snippet[]>('/snippets')

export const getSnippet = async (id: number): Promise<Snippet> => {
  if (!Number.isInteger(id) || id < 1) throw new NoRecordError()

  return apiClient.get<Snippet>(`/snippets/${id}`)
}

export const createSnippet = async (draft: SnippetDraft): Promise<Snippet> => {
  if (!sessionStore.read()) throw new UnauthenticatedError()
  validateSnippetDraft(draft).assertValid()

  return apiClient.post<Snippet>('/snippets', { body: draft, auth: true })
}
