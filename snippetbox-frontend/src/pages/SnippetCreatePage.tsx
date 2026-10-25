import { useNavigate } from 'react-router-dom'
import { useFlash } from '@/features/flash'
import { SnippetForm, type Snippet } from '@/features/snippets'

export function SnippetCreatePage() {
  const navigate = useNavigate()
  const flash = useFlash()

  const handleCreated = (snippet: Snippet) => {
    flash.push('Snippet successfully created!')
    navigate(`/snippet/view/${snippet.id}`)
  }

  return <SnippetForm onCreated={handleCreated} />
}
