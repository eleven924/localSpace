<template>
  <Transition name="modal">
    <div v-if="show" class="modal-overlay" @click="handleOverlayClick">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">{{ title }}</h3>
          <button v-if="closable" @click="cancel" class="modal-close" aria-label="Close">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>

        <div class="modal-body">
          <p class="modal-message">{{ message }}</p>
          <slot name="content"></slot>
        </div>

        <div class="modal-footer">
          <button
            type="button"
            @click="cancel"
            class="btn btn-secondary"
            :disabled="loading"
          >
            {{ cancelText }}
          </button>
          <button
            type="button"
            @click="confirm"
            class="btn"
            :class="confirmVariant"
            :disabled="loading"
          >
            <LoadingSpinner v-if="loading" :size="'small'" :inline="true" />
            <span v-else>{{ confirmText }}</span>
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import LoadingSpinner from './LoadingSpinner.vue'

interface Props {
  show: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  confirmVariant?: 'primary' | 'danger' | 'warning'
  closable?: boolean
  closeOnOverlay?: boolean
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  confirmText: 'Confirm',
  cancelText: 'Cancel',
  confirmVariant: 'primary',
  closable: true,
  closeOnOverlay: true,
  loading: false
})

const emit = defineEmits<{
  'update:show': [value: boolean]
  confirm: []
  cancel: []
}>()

const confirm = () => {
  emit('confirm')
}

const cancel = () => {
  emit('cancel')
  emit('update:show', false)
}

const handleOverlayClick = () => {
  if (props.closeOnOverlay && props.closable) {
    cancel()
  }
}

// Close dialog on Escape key
watch(() => props.show, (show) => {
  if (show) {
    document.addEventListener('keydown', handleEscape)
  } else {
    document.removeEventListener('keydown', handleEscape)
  }
})

const handleEscape = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && props.closable) {
    cancel()
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.modal-content {
  background-color: var(--surface-color);
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
}

.modal-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: var(--text-color);
}

.modal-close {
  background: none;
  border: none;
  color: var(--text-color);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.modal-close:hover {
  background-color: var(--border-color);
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.modal-message {
  font-size: 15px;
  line-height: 1.6;
  margin: 0 0 16px 0;
  color: var(--text-color);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
}

/* Modal transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .modal-content,
.modal-leave-active .modal-content {
  transition: transform 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-content,
.modal-leave-to .modal-content {
  transform: scale(0.95) translateY(-10px);
}

/* Button variants */
.btn.danger {
  background-color: var(--error-color);
  color: white;
}

.btn.danger:hover {
  opacity: 0.9;
}

.btn.warning {
  background-color: var(--warning-color);
  color: white;
}

.btn.warning:hover {
  opacity: 0.9;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Responsive adjustments */
@media (max-width: 480px) {
  .modal-content {
    border-radius: 8px;
    max-height: 95vh;
  }

  .modal-header,
  .modal-body,
  .modal-footer {
    padding: 16px;
  }

  .modal-title {
    font-size: 16px;
  }

  .modal-message {
    font-size: 14px;
  }

  .modal-footer {
    flex-direction: column-reverse;
  }

  .modal-footer .btn {
    width: 100%;
  }
}

/* Reduced motion preference */
@media (prefers-reduced-motion: reduce) {
  .modal-enter-active,
  .modal-leave-active,
  .modal-enter-active .modal-content,
  .modal-leave-active .modal-content {
    transition: none;
  }
}
</style>