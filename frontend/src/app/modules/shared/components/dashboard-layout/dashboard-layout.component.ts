import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NavbarComponent } from '../navbar/navbar.component';
import { FooterComponent } from '../footer/footer.component';

@Component({
  selector: 'app-dashboard-layout',
  standalone: true,
  imports: [CommonModule, NavbarComponent, FooterComponent],
  template: `
    <div class="dashboard-layout">
      <app-navbar></app-navbar>
      <main class="main-content">
        <ng-content></ng-content>
      </main>
      <app-footer></app-footer>
    </div>
  `,
  styleUrl: './dashboard-layout.component.css'
})
export class DashboardLayoutComponent {}
