import { Component, input, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-auth-card',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './auth-card.component.html',
  styleUrl: './auth-card.component.css'
})
export class AuthCardComponent {
  title = input<string>(''); // Using input() for standalone component
  subtitle = input<string>('');
  showHeader = input<boolean>(true);
  showFooter = input<boolean>(false);
}
