import { createDiscreteApi } from 'naive-ui'

const { message } = createDiscreteApi(['message'])

export const notify = {
  success: (content: string) => message.success(content),
  error: (content: string) => message.error(content),
  warning: (content: string) => message.warning(content),
}
