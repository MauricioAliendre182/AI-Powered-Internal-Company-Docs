import { Component, Input, Output, EventEmitter, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-auth-button',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './auth-button.component.html',
  styleUrl: './auth-button.component.css'
})
export class AuthButtonComponent {
  text = input<string>('Submit');
  type = input<'button' | 'submit' | 'reset'>('button');
  variant = input<'primary' | 'secondary' | 'outline' | 'ghost'>('primary');
  size = input<'sm' | 'md' | 'lg'>('md');
  icon = input<string>('');
  iconPosition = input<'left' | 'right'>('left');
  loading = input<boolean>(false);
  disabled = input<boolean>(false);
  fullWidth = input<boolean>(false);

  clicked = output<Event>();

  onClick(event: Event): void {
    if (!this.disabled() && !this.loading()) {
      this.clicked.emit(event);
    }
  }

  get buttonClasses(): string {
    const classes = [
      'auth-button',
      `variant-${this.variant()}`,
      `size-${this.size()}`
    ];

    if (this.loading()) classes.push('loading');
    if (this.disabled()) classes.push('disabled');
    if (this.fullWidth()) classes.push('full-width');
    if (this.icon()) classes.push('has-icon');

    return classes.join(' ');
  }
}
