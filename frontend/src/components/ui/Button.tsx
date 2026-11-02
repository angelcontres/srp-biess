import { forwardRef } from 'react';
import { LoaderCircle } from 'lucide-react';
import type { ButtonHTMLAttributes, ReactNode } from 'react';

export type ButtonVariant = 'primary' | 'secondary' | 'tertiary' | 'danger';
export type ButtonSize = 'sm' | 'md' | 'lg';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  /** All sizes keep the 44 px control height; only horizontal padding changes. */
  size?: ButtonSize;
  /** Blocks further clicks and swaps the label for a loader while keeping the width stable. */
  isLoading?: boolean;
  /** Text shown while `isLoading` is true, for example "Guardando…". */
  loadingText?: string;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  fullWidth?: boolean;
}

const variantClasses: Record<ButtonVariant, string> = {
  primary: 'border border-transparent bg-primary text-white hover:bg-primary-dark',
  secondary: 'border border-primary bg-surface text-primary hover:bg-info-50',
  tertiary: 'text-primary underline underline-offset-4 hover:text-primary-dark',
  danger: 'border border-transparent bg-danger text-white hover:brightness-90',
};

const sizeClasses: Record<ButtonSize, string> = {
  sm: 'h-11 px-4',
  md: 'h-11 px-5',
  lg: 'h-11 px-6',
};

const focusClasses =
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark';

const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button({
  variant = 'primary',
  size = 'md',
  isLoading = false,
  loadingText,
  leftIcon,
  rightIcon,
  fullWidth = false,
  disabled = false,
  className = '',
  children,
  type = 'button',
  onClick,
  ...rest
}, ref) {
  const paddingClasses = variant === 'tertiary' ? 'h-auto px-0' : sizeClasses[size];

  return (
    <button
      ref={ref}
      type={type}
      disabled={disabled}
      aria-disabled={isLoading || undefined}
      onClick={(event) => {
        if (isLoading) {
          event.preventDefault();
          return;
        }
        onClick?.(event);
      }}
      className={[
        'inline-flex items-center justify-center gap-2 rounded-md text-base font-semibold',
        variantClasses[variant],
        paddingClasses,
        focusClasses,
        'disabled:border disabled:border-line-input disabled:bg-neutral-100 disabled:text-muted disabled:no-underline',
        fullWidth ? 'w-full' : '',
        className,
      ]
        .filter(Boolean)
        .join(' ')}
      {...rest}
    >
      {isLoading ? (
        <span className="inline-grid">
          <span className="invisible col-start-1 row-start-1 inline-flex items-center gap-2">
            {children}
          </span>
          <span className="col-start-1 row-start-1 inline-flex items-center justify-center gap-2 whitespace-nowrap">
            <LoaderCircle
              size={20}
              strokeWidth={2}
              aria-hidden="true"
              className="shrink-0 motion-safe:animate-spin"
            />
            {loadingText ?? children}
          </span>
        </span>
      ) : (
        <>
          {leftIcon}
          <span>{children}</span>
          {rightIcon}
        </>
      )}
    </button>
  );
});

export default Button;
