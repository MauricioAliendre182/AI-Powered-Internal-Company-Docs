import { HttpClient, HttpHeaders } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '@environments/environment';
import {
  Document,
  DocumentUploadResponse,
  DocumentsListResponse,
  DocumentChunksResponse,
  QueryRequest,
  QueryResponse,
  DeleteDocumentResponse,
} from '@models/document.model';
import { checkToken } from '@interceptors/auth.interceptor';

@Injectable({
  providedIn: 'root',
})
export class DocumentService {
  private readonly apiUrl = environment.API_URL;
  private readonly http = inject(HttpClient);

  /**
   * Upload a document to the server
   * @param document FormData containing the file
   * @returns Observable with upload response
   */
  uploadDocument(document: FormData): Observable<DocumentUploadResponse> {
    return this.http.post<DocumentUploadResponse>(
      `${this.apiUrl}/documents`,
      document,
      {
        context: checkToken(),
      }
    );
  }

  /**
   * Get all documents
   * @returns Observable with list of documents
   */
  getAllDocuments(): Observable<DocumentsListResponse> {
    return this.http.get<DocumentsListResponse>(`${this.apiUrl}/documents`, {
      context: checkToken(),
    });
  }

  /**
   * Get chunks for a specific document
   * @param documentId The document ID
   * @returns Observable with document chunks
   */
  getDocumentChunks(documentId: string): Observable<DocumentChunksResponse> {
    return this.http.get<DocumentChunksResponse>(
      `${this.apiUrl}/documents/${documentId}/chunks`,
      {
        context: checkToken(),
      }
    );
  }

  /**
   * Delete a document
   * @param documentId The document ID to delete
   * @returns Observable with deletion response
   */
  deleteDocument(documentId: string): Observable<DeleteDocumentResponse> {
    return this.http.delete<DeleteDocumentResponse>(
      `${this.apiUrl}/documents/${documentId}`,
      {
        context: checkToken(),
      }
    );
  }

  /**
   * Query documents using RAG (Retrieval-Augmented Generation)
   * @param queryRequest The question to ask
   * @returns Observable with the AI-generated answer
   */
  queryDocuments(queryRequest: QueryRequest): Observable<QueryResponse> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
    });

    return this.http.post<QueryResponse>(`${this.apiUrl}/query`, queryRequest, {
      headers,
      context: checkToken(),
    });
  }

  /**
   * Convenience method to query documents with just a question string
   * @param question The question to ask
   * @returns Observable with the AI-generated answer
   */
  askQuestion(question: string): Observable<QueryResponse> {
    return this.queryDocuments({ question });
  }

  /**
   * Create FormData for file upload
   * @param file The file to upload
   * @param additionalData Optional additional form data
   * @returns FormData ready for upload
   */
  createFormData(
    file: File,
    additionalData?: Record<string, string>
  ): FormData {
    const formData = new FormData();
    formData.append('file', file, file.name);

    if (additionalData) {
      Object.keys(additionalData).forEach((key) => {
        formData.append(key, additionalData[key]);
      });
    }

    return formData;
  }

  /**
   * Validate file before upload
   * @param file The file to validate
   * @returns Validation result
   */
  validateFile(file: File): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];
    const maxSize = 10 * 1024 * 1024; // 10MB
    const allowedTypes = [
      'application/pdf',
      'text/plain',
      'application/msword',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'text/csv',
      'application/vnd.ms-excel',
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    ];

    // Check file size
    if (file.size > maxSize) {
      errors.push('File size must be less than 10MB');
    }

    // Check file type
    if (!allowedTypes.includes(file.type)) {
      errors.push(
        'File type not supported. Please upload PDF, DOC, DOCX, TXT, CSV, XLS, or XLSX files.'
      );
    }

    // Check if file name is provided
    if (!file.name || file.name.trim() === '') {
      errors.push('File must have a valid name');
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  }

  /**
   * Format file size for display
   * @param bytes File size in bytes
   * @returns Formatted file size string
   */
  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  /**
   * Get file icon based on file type
   * @param mimeType The MIME type of the file
   * @returns FontAwesome icon class
   */
  getFileIcon(mimeType: string): string {
    const iconMap: Record<string, string> = {
      'application/pdf': 'fas fa-file-pdf',
      'text/plain': 'fas fa-file-alt',
      'application/msword': 'fas fa-file-word',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document':
        'fas fa-file-word',
      'text/csv': 'fas fa-file-csv',
      'application/vnd.ms-excel': 'fas fa-file-excel',
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet':
        'fas fa-file-excel',
      'image/jpeg': 'fas fa-file-image',
      'image/png': 'fas fa-file-image',
      'image/gif': 'fas fa-file-image',
    };

    return iconMap[mimeType] || 'fas fa-file';
  }
}
