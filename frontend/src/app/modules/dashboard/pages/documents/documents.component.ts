import { Component, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ListDocumentsComponent } from '@shared/components/list-documents/list-documents.component';
import { DashboardLayoutComponent } from '@shared/components/dashboard-layout/dashboard-layout.component';
import { DocumentUploadService } from '@services/document-upload.service';

@Component({
  selector: 'app-documents',
  standalone: true,
  imports: [CommonModule, ListDocumentsComponent, DashboardLayoutComponent],
  templateUrl: './documents.component.html',
  styleUrl: './documents.component.css'
})
export class DocumentsComponent {
  private readonly documentUploadService = inject(DocumentUploadService);

  // Signal to trigger document reload in the list component
  receivedSignalToLoadDocumentsOnUpload = signal<boolean>(false);

  constructor() {
    // Effect to watch for upload completion from shared service
    effect(() => {
      const uploadCompleted = this.documentUploadService.uploadSignal();
      if (uploadCompleted) {
        console.log('Upload detected from shared service, triggering reload');
        this.triggerDocumentReload();
      }
    });
  }

  // Method to trigger document reload (can be called from other components or events)
  triggerDocumentReload(): void {
    this.receivedSignalToLoadDocumentsOnUpload.set(true);
    // Reset the signal after a short delay to allow for future triggers
    setTimeout(() => {
      this.documentUploadService.setUploadSignal(false);
      console.log('Document reload signal reset');
      this.receivedSignalToLoadDocumentsOnUpload.set(false);
    }, 100);
  }

  // Handle the event when documents are uploaded
  loadDocumentsOnUpload(event: any): void {
    console.log('Documents uploaded:', event);
    this.triggerDocumentReload();
  }

  // Method to get the current upload signal value
  getUploadSignalValue(): boolean {
    return this.documentUploadService.uploadSignal();
  }
}
