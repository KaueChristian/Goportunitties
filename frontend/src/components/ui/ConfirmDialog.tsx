import styles from './ConfirmDialog.module.css'
import { Modal } from './Modal'
import { Button } from './Button'
import { Icon } from './Icon'

interface ConfirmDialogProps {
  open: boolean
  title: string
  message: string
  confirmLabel?: string
  loading?: boolean
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = 'Confirmar',
  loading = false,
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  return (
    <Modal
      open={open}
      title={title}
      onClose={onCancel}
      footer={
        <>
          <Button variant="ghost" onClick={onCancel} disabled={loading}>
            Cancelar
          </Button>
          <Button variant="danger" onClick={onConfirm} loading={loading}>
            {confirmLabel}
          </Button>
        </>
      }
    >
      <div className={styles.body}>
        <div className={styles.iconRing}>
          <Icon name="alert" size={22} />
        </div>
        <p className={styles.message}>{message}</p>
      </div>
    </Modal>
  )
}
