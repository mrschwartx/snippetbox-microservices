export interface Snippet {
  id: number
  title: string
  content: string
  created: string
  expires: string
}

export interface SnippetDraft {
  title: string
  content: string
  expires: number
}
