import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles/global.css'
import App from './App.tsx'
import { ToastProvider } from './components/ui/Toast.tsx'
import { OpeningsProvider } from './state/OpeningsProvider.tsx'
import { DialogsProvider } from './state/DialogsProvider.tsx'

// Order matters: dialogs report their outcome through toasts and write through
// the openings store, so both have to be above them.
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ToastProvider>
      <OpeningsProvider>
        <DialogsProvider>
          <App />
        </DialogsProvider>
      </OpeningsProvider>
    </ToastProvider>
  </StrictMode>,
)
