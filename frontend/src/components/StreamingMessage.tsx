import { Typography } from 'antd'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'

const { Text } = Typography

interface Props {
  content: string
  isStreaming?: boolean
}

export default function StreamingMessage({ content, isStreaming }: Props) {
  if (!content) {
    return isStreaming ? <Text className="streaming-cursor" type="secondary">思考中</Text> : null
  }

  return (
    <div className="markdown-body">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeHighlight]}
        components={{
          pre: ({ children }) => (
            <pre style={{ borderRadius: 8, overflow: 'auto', padding: 12, background: '#1e1e1e' }}>
              {children}
            </pre>
          ),
          code: ({ children, className }) => {
            const isInline = !className
            return isInline ? (
              <code style={{ background: '#f0f0f0', padding: '2px 6px', borderRadius: 4, fontSize: '0.9em' }}>
                {children}
              </code>
            ) : (
              <code className={className}>{children}</code>
            )
          },
        }}
      >{content}</ReactMarkdown>
      {isStreaming && <span className="streaming-cursor" />}
    </div>
  )
}
