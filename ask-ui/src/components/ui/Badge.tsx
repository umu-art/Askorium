import { type HTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

type Variant = 'default' | 'brand' | 'muted'

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: Variant
}

const variantClasses: Record<Variant, string> = {
  default: 'bg-gray-100 text-gray-600',
  brand: 'bg-brand-50 text-brand-600',
  muted: 'bg-gray-50 text-gray-400 border border-gray-100',
}

export function Badge({ className, variant = 'default', children, ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
        variantClasses[variant],
        className
      )}
      {...props}
    >
      {children}
    </span>
  )
}
