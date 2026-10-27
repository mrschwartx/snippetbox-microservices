interface SubmitButtonProps {
  label: string
  isPending: boolean
}

export function SubmitButton({ label, isPending }: SubmitButtonProps) {
  return (
    <div className="border-t border-dashed border-line">
      <input
        type="submit"
        value={label}
        disabled={isPending}
        className="mt-[1rem] inline-block cursor-pointer rounded-[3px] bg-accent px-[1.5rem] py-[1rem] font-bold text-white hover:bg-accent-hover disabled:cursor-progress disabled:opacity-60"
      />
    </div>
  )
}
