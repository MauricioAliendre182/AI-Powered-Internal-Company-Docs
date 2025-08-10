import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { QuerySectionComponent } from '@shared/components/query-section/query-section.component';
import { DashboardLayoutComponent } from '@shared/components/dashboard-layout/dashboard-layout.component';

@Component({
  selector: 'app-query',
  standalone: true,
  imports: [CommonModule, QuerySectionComponent, DashboardLayoutComponent],
  templateUrl: './query.component.html',
  styleUrl: './query.component.css'
})
export class QueryComponent {

}
