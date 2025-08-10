import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { QueryResponse } from '@models/document.model';
import { RequestStatus } from '@models/request-status.model';
import { DocumentService } from '@services/document.service';

@Component({
  selector: 'app-query-section',
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './query-section.component.html',
  styleUrl: './query-section.component.css'
})
export class QuerySectionComponent {
  private readonly documentService = inject(DocumentService);
  private readonly fb = inject(FormBuilder);

  // Query properties
  queryForm: FormGroup;
  queryResponse: QueryResponse | null = null;
  queryStatus: RequestStatus = 'init';
  queryError = '';

  constructor() {
    this.queryForm = this.fb.group({
      question: ['', [Validators.required, Validators.minLength(3)]]
    });
  }

  // Query methods
  queryDocuments(): void {
    if (this.queryForm.invalid) return;

    this.queryStatus = 'loading';
    this.queryError = '';

    // Get the question from the form
    // It retrieves the question from the form control
    const question = this.queryForm.get('question')?.value;

    // Call the document service to ask a question
    // It subscribes to the observable returned by the service
    this.documentService.askQuestion(question).subscribe({
      next: (response) => {
        this.queryResponse = response;
        this.queryStatus = 'success';
      },
      error: (error) => {
        this.queryStatus = 'failed';
        this.queryError = error.error?.error || 'Failed to get AI response';
        console.error('Query failed:', error);
      }
    });
  }
}
