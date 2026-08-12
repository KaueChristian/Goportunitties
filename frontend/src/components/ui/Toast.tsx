import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import styles from './Toast.module.css'
import { Icon } from './Icon'

type ToastTone = 'success' | 'error'

interface Toast {
  id: number
  tone: ToastTone
  message: string
}

interface ToastContextValue {
  notifySuccess: (message: string) => void
  notifyError: (message: string) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

const DISMISS_AFTER = 4000

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const dismiss = useCallback((id: number) => {
    setToasts((current) => current.filter((toast) => toast.id !== id))
  }, [])

  const push = useCallback(
    (tone: ToastTone, message: string) => {
      const id = Date.now() + Math.random()
      setToasts((current) => [...current, { id, tone, message }])
      window.setTimeout(() => dismiss(id), DISMISS_AFTER)
    },
    [dismiss],
  )

  const value = useMemo<ToastContextValue>(
    () => ({
      notifySuccess: (message: string) => push('success', message),
      notifyError: (message: string) => push('error', message),
    }),
    [push],
  )

  return (
    <ToastContext.Provider value={value}>
      {children}
      {createPortal(
        <div className={styles.stack} role="region" aria-label="Notificações">
          {toasts.map((toast) => (
            <output key={toast.id} className={`${styles.toast} ${styles[toast.tone]}`}>
              <Icon name={toast.tone === 'success' ? 'check' : 'alert'} size={16} />
              <span className={styles.message}>{toast.message}</span>
              <button
                className={styles.close}
                onClick={() => dismiss(toast.id)}
                aria-label="Dispensar"
              >
                <Icon name="close" size={14} />
              </button>
            </output>
          ))}
        </div>,
        document.body,
      )}
    </ToastContext.Provider>
  )
}

export function useToast(): ToastContextValue {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToast precisa estar dentro de <ToastProvider>')
  }
  return context
}
