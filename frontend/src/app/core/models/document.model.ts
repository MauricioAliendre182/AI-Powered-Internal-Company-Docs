export interface Document {
  id: string;
  name: string;
  originalFilename: string;
  uploadedAt: string;
  size?: number;
  contentType?: string;
}

export interface DocumentResponse {
  id: string;
  name: string;
  originalFilename: string;
  uploadedAt: string;
}

export interface DocumentUploadResponse {
  message: string;
  document: DocumentResponse;
  chunks_created?: number;
}

export interface DocumentChunk {
  id: string;
  documentId: string;
  content: string;
  chunk_index: number;
  size: number;
  contentType: string;
  createdAt: string;
}

export interface DocumentChunksResponse {
  documentId: string;
  chunks: DocumentChunk[];
}

export interface QueryRequest {
  question: string;
}

export interface QueryResponse {
  question: string;
  answer: string;
}

export interface DocumentsListResponse {
  documents: Document[];
}

export interface DeleteDocumentResponse {
  message: string;
}
