import { useState, type FormEvent } from 'react'
import { Shield, ArrowRight } from 'lucide-react'
import { Button } from './ui/Button'
import { useApp } from '../context/AppContext'

export function LoginScreen() {
  const { authError, login } = useApp()
  const [showForm, setShowForm] = useState(false)
  const [tailnet, setTailnet] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const trimmed = tailnet.trim()
    if (!trimmed) {
      setSubmitError('Tailnet is required.')
      return
    }

    setSubmitting(true)
    setSubmitError(null)
    try {
      await login(trimmed)
    } catch {
      setSubmitError('Could not connect this tailnet.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-full items-center justify-center bg-[radial-gradient(circle_at_top,_rgba(248,113,113,0.12),_transparent_42%),linear-gradient(180deg,#0f1115_0%,#090b0e_100%)] px-6 py-12">
      <div className="w-full max-w-md rounded-3xl border border-white/10 bg-white/5 p-8 shadow-2xl shadow-black/30 backdrop-blur">
        <div className="mb-6 inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-rose-500/15 text-rose-300">
          <Shield size={22} />
        </div>

        <h1 className="text-2xl font-semibold tracking-tight text-white">BridgeAI Beta</h1>
        <p className="mt-3 text-sm leading-6 text-slate-300">
          Connect your tailnet to load devices from Tailscale and merge them with live Bridge agents.
        </p>

        {!showForm ? (
          <div className="mt-8">
            <Button
              variant="primary"
              size="md"
              onClick={() => setShowForm(true)}
              className="h-11 w-full justify-center text-sm font-semibold"
            >
              Continue with Tailscale
            </Button>
          </div>
        ) : (
          <form className="mt-8 space-y-4" onSubmit={handleSubmit}>
            <div>
              <label htmlFor="tailnet" className="mb-2 block text-xs font-semibold uppercase tracking-[0.22em] text-slate-400">
                Tailnet
              </label>
              <input
                id="tailnet"
                value={tailnet}
                onChange={(event) => setTailnet(event.target.value)}
                placeholder="example.ts.net"
                autoComplete="off"
                className="h-11 w-full rounded-2xl border border-white/10 bg-black/20 px-4 text-sm text-white outline-none transition focus:border-rose-400/60"
              />
            </div>

            <Button
              type="submit"
              variant="primary"
              size="md"
              disabled={submitting}
              className="h-11 w-full justify-center text-sm font-semibold"
            >
              {submitting ? 'Connecting…' : (
                <>
                  Connect
                  <ArrowRight size={15} />
                </>
              )}
            </Button>
          </form>
        )}

        {(submitError || authError) && (
          <div className="mt-5 rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200">
            {submitError ?? authError}
          </div>
        )}

        <p className="mt-5 text-xs leading-5 text-slate-400">
          Enter the tailnet name you want to inspect.
        </p>
      </div>
    </div>
  )
}
