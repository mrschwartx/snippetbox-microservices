import { useParams } from 'react-router-dom'
import { ErrorAlert } from '@/components/ui/ErrorAlert'
import { Loading } from '@/components/ui/Loading'
import { SnippetCard, parseSnippetId, useSnippet } from '@/features/snippets'
import { NoRecordError, UNEXPECTED_ERROR_MESSAGE } from '@/lib/errors'

export function SnippetViewPage() {
  const { id } = useParams<{ id: string }>()
  const { data: snippet, error, isPending } = useSnippet(parseSnippetId(id))

  if (isPending) return <Loading />
  if (error instanceof NoRecordError) return <p>Not Found</p>
  if (error) return <ErrorAlert message={UNEXPECTED_ERROR_MESSAGE} />
  return <SnippetCard snippet={snippet} />
}
