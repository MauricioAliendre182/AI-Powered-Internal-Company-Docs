import { Component, inject, output } from '@angular/core';
import { RequestStatus } from '@models/request-status.model';
import { DocumentService } from '@services/document.service';

@Component({
  selector: 'app-upload-document',
  imports: [],
  templateUrl: './upload-document.component.html',
  styleUrl: './upload-document.component.css'
})
export class UploadDocumentComponent {
  private readonly documentService = inject(DocumentService);

  // File upload properties
  selectedFile: File | null = null;
  isDragOver = false;
  validationErrors: string[] = [];
  uploadStatus: RequestStatus = 'init';
  uploadError = '';

  // Output signal to update the documents list
  loadDocumentsOnUpload = output<boolean>();

  // File upload methods
  // This method handles drag and drop events for file uploads
  // It prevents the default behavior and sets the drag over state
  onDragOver(event: DragEvent): void {
    // event.preventDefault() is used to prevent the browser's default handling of the event
    // It also checks if the dragged item is a file
    event.preventDefault();
    this.isDragOver = true;
  }

  // This method handles drag leave events for file uploads
  // It prevents the default behavior and resets the drag over state
  // It also checks if the dragged item is a file
  onDragLeave(event: DragEvent): void {
    // event.preventDefault() is used to prevent the browser's default handling of the event
    // It also checks if the dragged item is a file
    event.preventDefault();
    this.isDragOver = false;
  }

  // This method handles drop events for file uploads
  // It prevents the default behavior, resets the drag over state, and processes the dropped file
  onDrop(event: DragEvent): void {
    event.preventDefault();
    this.isDragOver = false;

    // event.dataTransfer?.files is used to get the files dropped
    // It checks if there are files and processes the first one
    const files = event.dataTransfer?.files;
    if (files && files.length > 0) {
      // Call the method to handle file selection
      // This method will validate the file and set it for upload
      this.handleFileSelection(files[0]);
    }
  }

  // This method handles file selection from an input element
  // It checks if files are selected and processes the first one
  // It also validates the file and sets it for upload
  // This method is called when the user selects a file from the file input
  onFileSelected(event: Event): void {
    // event.target is used to get the input element
    // It checks if files are selected and processes the first one
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      // Call the method to handle file selection
      // This method will validate the file and set it for upload
      this.handleFileSelection(input.files[0]);
    }
  }

  // This method handles the file selection logic
  // It sets the selected file, resets the upload status, and validates the file
  private handleFileSelection(file: File): void {
    this.selectedFile = file;
    this.uploadStatus = 'init';
    this.uploadError = '';

    // Validate the file using the document service
    // This method will check the file type, size, and other criteria
    const validation = this.documentService.validateFile(file);
    this.validationErrors = validation.errors;
  }

  // This method removes the selected file and resets the upload state
  // It clears the selected file, validation errors, upload status, and error message
  removeFile(): void {
    this.selectedFile = null;
    this.validationErrors = [];
    this.uploadStatus = 'init';
    this.uploadError = '';
  }

  // This method uploads the selected file
  // It creates a FormData object, calls the document service to upload the file,
  uploadDocument(): void {
    // Check if a file is selected and if there are validation errors
    // If not, it returns early to prevent unnecessary uploads
    if (!this.selectedFile || this.validationErrors.length > 0) return;

    // Set the upload status to loading to indicate the upload is in progress
    // This will trigger any loading indicators in the UI
    this.uploadStatus = 'loading';

    // Create FormData for the file upload
    // This method creates a FormData object with the selected file
    const formData = this.documentService.createFormData(this.selectedFile);

    // Call the document service to upload the document
    // It subscribes to the observable returned by the service
    this.documentService.uploadDocument(formData).subscribe({
      next: (response) => {
        this.uploadStatus = 'success';
        console.log('Document uploaded:', response);
        // Send a notification or update the UI as needed
        // After successful upload, reload the documents list
        this.loadDocumentsOnUpload.emit(true);

        // Reset form after successful upload
        setTimeout(() => {
          this.removeFile();
          this.uploadStatus = 'init';
        }, 2000);
      },
      error: (error) => {
        this.uploadStatus = 'failed';
        this.uploadError = error.error?.error || 'Failed to upload document';
        console.error('Upload failed:', error);
      }
    });
  }

  // Utility methods
  // These methods provide utility functions for the component
  // this method gets the file icon based on the MIME type
  // It uses the document service to get the appropriate icon for the file type
  getFileIcon(mimeType: string): string {
    return this.documentService.getFileIcon(mimeType);
  }

  // This method formats the file size for display
  // It uses the document service to format the file size in a human-readable way
  formatFileSize(bytes: number): string {
    return this.documentService.formatFileSize(bytes);
  }

  // This method formats the date for display
  // It uses the document service to format the date in a human-readable way
  formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
}
