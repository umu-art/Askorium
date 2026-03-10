import { useState } from 'react'
import {
  Globe, Plus, RefreshCw, Trash2, Pencil, Zap,
  Clock, ExternalLink, AlertCircle, RotateCcw,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { LogoIcon } from '@/components/icons/LogoIcon'
import { cn, extractDomain } from '@/lib/utils'
import { useSources } from '@/hooks/useSources'
import type { SourceDto, SourceAutoSyncPolicy } from '@/lib/api'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function relativeTime(date: Date | undefined): string {
  if (!date) return 'никогда'
  const diff = Date.now() - date.getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'только что'
  if (mins < 60) return `${mins} мин. назад`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} ч. назад`
  const days = Math.floor(hours / 24)
  return `${days} дн. назад`
}

function syncStatusColor(policy?: SourceAutoSyncPolicy): string {
  if (!policy?.enabled) return 'bg-gray-300'
  if (!policy.lastSyncedAt) return 'bg-amber-400'
  const hoursSince = (Date.now() - policy.lastSyncedAt.getTime()) / 3_600_000
  if (hoursSince < policy.intervalMinutes / 60) return 'bg-emerald-400'
  return 'bg-amber-400'
}

// ---------------------------------------------------------------------------
// Modal
// ---------------------------------------------------------------------------

function Modal({ open, onClose, children }: {
  open: boolean
  onClose: () => void
  children: React.ReactNode
}) {
  if (!open) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-gray-900/40 backdrop-blur-sm animate-fade-in"
        onClick={onClose}
      />
      <div className="relative z-10 w-full max-w-lg mx-4 animate-slide-up">
        {children}
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Source Form
// ---------------------------------------------------------------------------

interface SourceFormProps {
  initial?: SourceDto
  onSubmit: (dto: SourceDto) => Promise<void>
  onCancel: () => void
}

function SourceForm({ initial, onSubmit, onCancel }: SourceFormProps) {
  const [url, setUrl] = useState(initial?.sourceUrl ?? '')
  const [autoSync, setAutoSync] = useState(initial?.syncPolicy?.enabled ?? false)
  const [interval, setInterval] = useState(initial?.syncPolicy?.intervalMinutes ?? 720)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isEdit = !!initial?.id

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!url.trim()) return

    const dto: SourceDto = {
      ...(initial?.id ? { id: initial.id } : {}),
      sourceUrl: url.trim(),
      syncPolicy: {
        enabled: autoSync,
        intervalMinutes: interval,
        ...(initial?.syncPolicy?.lastSyncedAt
          ? { lastSyncedAt: initial.syncPolicy.lastSyncedAt }
          : {}),
      },
    }

    setSaving(true)
    setError(null)
    try {
      await onSubmit(dto)
    } catch {
      setError('Не удалось сохранить источник')
      setSaving(false)
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="rounded-2xl border border-gray-200 bg-white p-6 shadow-xl shadow-gray-200/60"
    >
      <h2 className="text-lg font-semibold text-gray-900 mb-5">
        {isEdit ? 'Редактирование источника' : 'Новый источник'}
      </h2>

      {/* URL */}
      <label className="block mb-4">
        <span className="text-xs font-medium uppercase tracking-wide text-gray-500 mb-1.5 block">
          URL сайта
        </span>
        <input
          type="url"
          value={url}
          onChange={e => setUrl(e.target.value)}
          placeholder="https://example.com"
          required
          className={cn(
            'w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-900',
            'placeholder:text-gray-400',
            'focus:border-brand-300 focus:bg-white focus:outline-none focus:ring-2 focus:ring-brand-100',
            'transition-colors',
          )}
        />
      </label>

      {/* Auto-sync toggle */}
      <div className="flex items-center justify-between mb-4 py-3 px-4 rounded-xl bg-gray-50 border border-gray-100">
        <div>
          <p className="text-sm font-medium text-gray-700">Авто-синхронизация</p>
          <p className="text-xs text-gray-400 mt-0.5">Периодически обновлять индекс</p>
        </div>
        <button
          type="button"
          role="switch"
          aria-checked={autoSync}
          onClick={() => setAutoSync(!autoSync)}
          className={cn(
            'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full transition-colors',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-400 focus-visible:ring-offset-2',
            autoSync ? 'bg-brand' : 'bg-gray-300',
          )}
        >
          <span
            className={cn(
              'pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-sm ring-0 transition-transform',
              autoSync ? 'translate-x-[22px]' : 'translate-x-0.5',
              'mt-0.5',
            )}
          />
        </button>
      </div>

      {/* Interval */}
      {autoSync && (
        <label className="block mb-5 animate-fade-in">
          <span className="text-xs font-medium uppercase tracking-wide text-gray-500 mb-1.5 block">
            Интервал (минуты)
          </span>
          <input
            type="number"
            min={10}
            value={interval}
            onChange={e => setInterval(Number(e.target.value))}
            className={cn(
              'w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-900',
              'focus:border-brand-300 focus:bg-white focus:outline-none focus:ring-2 focus:ring-brand-100',
              'transition-colors',
            )}
          />
          <p className="mt-1.5 text-xs text-gray-400">
            Минимум 10 мин. &middot; {interval >= 60
              ? `${Math.floor(interval / 60)} ч. ${interval % 60 ? `${interval % 60} мин.` : ''}`
              : `${interval} мин.`}
          </p>
        </label>
      )}

      {error && (
        <div className="mb-4 flex items-center gap-2 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">
          <AlertCircle className="h-4 w-4 shrink-0" />
          {error}
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-3 justify-end">
        <Button variant="secondary" type="button" onClick={onCancel} disabled={saving}>
          Отмена
        </Button>
        <Button type="submit" isLoading={saving}>
          {isEdit ? 'Сохранить' : 'Добавить'}
        </Button>
      </div>
    </form>
  )
}

// ---------------------------------------------------------------------------
// Delete confirm
// ---------------------------------------------------------------------------

function DeleteConfirm({ source, onConfirm, onCancel }: {
  source: SourceDto
  onConfirm: () => Promise<void>
  onCancel: () => void
}) {
  const [deleting, setDeleting] = useState(false)

  async function handleDelete() {
    setDeleting(true)
    try {
      await onConfirm()
    } catch {
      setDeleting(false)
    }
  }

  return (
    <div className="rounded-2xl border border-gray-200 bg-white p-6 shadow-xl shadow-gray-200/60">
      <div className="flex items-center gap-3 mb-4">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-100">
          <Trash2 className="h-5 w-5 text-red-600" />
        </div>
        <div>
          <h2 className="text-lg font-semibold text-gray-900">Удалить источник?</h2>
          <p className="text-sm text-gray-500">Это действие нельзя отменить</p>
        </div>
      </div>

      <div className="mb-5 rounded-xl bg-gray-50 border border-gray-100 px-4 py-3">
        <p className="text-sm font-medium text-gray-700 break-all">{source.sourceUrl}</p>
        {source.id && (
          <p className="text-xs text-gray-400 mt-1 font-mono">{source.id}</p>
        )}
      </div>

      <div className="flex gap-3 justify-end">
        <Button variant="secondary" onClick={onCancel} disabled={deleting}>
          Отмена
        </Button>
        <button
          onClick={handleDelete}
          disabled={deleting}
          className={cn(
            'inline-flex items-center gap-2 rounded-xl px-4 h-10 text-sm font-medium',
            'bg-red-600 text-white shadow-sm',
            'hover:bg-red-700 active:bg-red-800',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-400 focus-visible:ring-offset-2',
            'disabled:pointer-events-none disabled:opacity-40',
            'transition-colors',
          )}
        >
          {deleting && <Spinner size="sm" className="border-red-300 border-t-white" />}
          Удалить
        </button>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Source card
// ---------------------------------------------------------------------------

function SourceCard({ source, isSyncing, onEdit, onDelete, onSync }: {
  source: SourceDto
  isSyncing: boolean
  onEdit: () => void
  onDelete: () => void
  onSync: (force: boolean) => void
}) {
  const policy = source.syncPolicy
  const domain = extractDomain(source.sourceUrl)

  return (
    <div
      className={cn(
        'group relative rounded-2xl border bg-white p-5',
        'transition-all duration-200',
        'hover:shadow-md hover:shadow-gray-100 hover:border-gray-200',
        'border-gray-100',
      )}
    >
      {/* Status dot */}
      <div className={cn(
        'absolute top-5 right-5 h-2.5 w-2.5 rounded-full ring-2 ring-white',
        syncStatusColor(policy),
        isSyncing && 'animate-pulse',
      )} />

      {/* URL */}
      <div className="pr-8 mb-3">
        <a
          href={source.sourceUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-gray-900 hover:text-brand-600 transition-colors"
        >
          <Globe className="h-4 w-4 text-gray-400 shrink-0" />
          {domain}
          <ExternalLink className="h-3 w-3 text-gray-300" />
        </a>
        <p className="text-xs text-gray-400 mt-0.5 break-all">{source.sourceUrl}</p>
      </div>

      {/* Meta row */}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500 mb-4">
        {policy?.enabled ? (
          <span className="inline-flex items-center gap-1">
            <RefreshCw className="h-3 w-3" />
            каждые {policy.intervalMinutes >= 60
              ? `${Math.floor(policy.intervalMinutes / 60)} ч.`
              : `${policy.intervalMinutes} мин.`}
          </span>
        ) : (
          <span className="text-gray-400">авто-синхр. выкл.</span>
        )}

        <span className="inline-flex items-center gap-1">
          <Clock className="h-3 w-3" />
          {relativeTime(policy?.lastSyncedAt)}
        </span>

        {source.id && (
          <span className="font-mono text-gray-300 hidden sm:inline">{source.id.slice(0, 8)}</span>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2">
        <Button
          variant="secondary"
          size="sm"
          onClick={() => onSync(false)}
          isLoading={isSyncing}
          disabled={isSyncing}
        >
          <RefreshCw className="h-3.5 w-3.5 mr-1" />
          Синхр.
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onSync(true)}
          disabled={isSyncing}
        >
          <Zap className="h-3.5 w-3.5 mr-1" />
          Полная
        </Button>

        <div className="flex-1" />

        <button
          onClick={onEdit}
          className="p-2 rounded-lg text-gray-400 hover:text-brand-600 hover:bg-brand-50 transition-colors"
          aria-label="Редактировать"
        >
          <Pencil className="h-4 w-4" />
        </button>
        <button
          onClick={onDelete}
          className="p-2 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors"
          aria-label="Удалить"
        >
          <Trash2 className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Empty state
// ---------------------------------------------------------------------------

function EmptyState({ onAdd }: { onAdd: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
      <div className="mb-5 flex h-16 w-16 items-center justify-center rounded-2xl bg-brand-50">
        <Globe className="h-8 w-8 text-brand-400" />
      </div>
      <h3 className="text-base font-semibold text-gray-900 mb-1">Нет источников</h3>
      <p className="text-sm text-gray-500 mb-6 text-center max-w-xs">
        Добавьте URL сайта, чтобы начать индексацию и поиск по его содержимому
      </p>
      <Button onClick={onAdd}>
        <Plus className="h-4 w-4 mr-1.5" />
        Добавить источник
      </Button>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

type ModalState =
  | { type: 'closed' }
  | { type: 'add' }
  | { type: 'edit'; source: SourceDto }
  | { type: 'delete'; source: SourceDto }

export function SourcesPage() {
  const {
    sources, isLoading, error,
    syncingIds, isAutoSyncing,
    fetchSources, upsertSource, deleteSource, syncSource, triggerAutoSync,
  } = useSources()

  const [modal, setModal] = useState<ModalState>({ type: 'closed' })

  async function handleUpsert(dto: SourceDto) {
    await upsertSource(dto)
    setModal({ type: 'closed' })
  }

  async function handleDelete(sourceId: string) {
    await deleteSource(sourceId)
    setModal({ type: 'closed' })
  }

  return (
    <div className="min-h-screen bg-[#fafaf9]">
      {/* ── Header ─────────────────────────────────────────────────── */}
      <header className="sticky top-0 z-10 border-b border-gray-100 bg-white/90 backdrop-blur-sm">
        <div className="mx-auto flex max-w-4xl items-center gap-3 px-6 h-14">
          <LogoIcon className="h-7 w-7 rounded-lg shrink-0" />
          <h1 className="text-base font-semibold text-gray-900 mr-auto">Источники</h1>

          <Button
            variant="ghost"
            size="sm"
            onClick={triggerAutoSync}
            isLoading={isAutoSyncing}
            disabled={isAutoSyncing || sources.length === 0}
          >
            <RotateCcw className="h-3.5 w-3.5 mr-1.5" />
            Авто-синхр.
          </Button>

          <Button size="sm" onClick={() => setModal({ type: 'add' })}>
            <Plus className="h-4 w-4 mr-1" />
            Добавить
          </Button>
        </div>
      </header>

      {/* ── Content ────────────────────────────────────────────────── */}
      <main className="mx-auto max-w-4xl px-6 py-8">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Spinner size="lg" />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
            <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100">
              <AlertCircle className="h-6 w-6 text-red-500" />
            </div>
            <p className="text-sm text-gray-600 mb-4">{error}</p>
            <Button variant="secondary" onClick={fetchSources}>
              Попробовать снова
            </Button>
          </div>
        ) : sources.length === 0 ? (
          <EmptyState onAdd={() => setModal({ type: 'add' })} />
        ) : (
          <div className="grid gap-4">
            {sources.map((source, i) => (
              <div
                key={source.id ?? i}
                className="animate-slide-up"
                style={{ animationDelay: `${i * 60}ms`, animationFillMode: 'both' }}
              >
                <SourceCard
                  source={source}
                  isSyncing={!!source.id && syncingIds.has(source.id)}
                  onEdit={() => setModal({ type: 'edit', source })}
                  onDelete={() => setModal({ type: 'delete', source })}
                  onSync={force => source.id && syncSource(source.id, force)}
                />
              </div>
            ))}
          </div>
        )}
      </main>

      {/* ── Modals ─────────────────────────────────────────────────── */}
      <Modal
        open={modal.type === 'add' || modal.type === 'edit'}
        onClose={() => setModal({ type: 'closed' })}
      >
        <SourceForm
          initial={modal.type === 'edit' ? modal.source : undefined}
          onSubmit={handleUpsert}
          onCancel={() => setModal({ type: 'closed' })}
        />
      </Modal>

      <Modal
        open={modal.type === 'delete'}
        onClose={() => setModal({ type: 'closed' })}
      >
        {modal.type === 'delete' && (
          <DeleteConfirm
            source={modal.source}
            onConfirm={() => handleDelete(modal.source.id!)}
            onCancel={() => setModal({ type: 'closed' })}
          />
        )}
      </Modal>
    </div>
  )
}
