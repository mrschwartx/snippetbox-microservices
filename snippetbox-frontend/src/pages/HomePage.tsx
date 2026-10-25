import { ErrorAlert } from '@/components/ui/ErrorAlert'
import { Loading } from '@/components/ui/Loading'
import { SnippetTable, useLatestSnippets } from '@/features/snippets'
import { UNEXPECTED_ERROR_MESSAGE } from '@/lib/errors'

export function HomePage() {
  const { data: snippets, isPending, isError } = useLatestSnippets()

  return (
    <>
      <h2 className="mb-9 text-[1.2rem] font-bold">Latest Snippets</h2>
      {isPending && <Loading />}
      {isError && <ErrorAlert message={UNEXPECTED_ERROR_MESSAGE} />}
      {snippets &&
        (snippets.length > 0 ? (
          <SnippetTable snippets={snippets} />
        ) : (
          <p>There&apos;s nothing to see here... yet!</p>
        ))}
    </>
  )
}
