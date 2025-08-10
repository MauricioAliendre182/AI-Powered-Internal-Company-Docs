-- Check for corrupted embeddings in the database
-- This query will show the length of embedding vectors and detect any that contain timestamps

SELECT 
    id,
    document_id,
    chunk_index,
    length(embedding::text) as embedding_text_length,
    substring(embedding::text, 1, 100) as embedding_preview,
    substring(content, 1, 100) as content_preview
FROM chunks
ORDER BY chunk_index
LIMIT 10;

-- Check for embeddings that contain timestamps (likely corrupted)
SELECT 
    id,
    document_id,
    chunk_index,
    content,
    embedding::text
FROM chunks
WHERE embedding::text LIKE '%2025-08-%'
LIMIT 5;

-- Delete corrupted chunks if any found
-- Uncomment this line only after confirming corrupted data exists:
-- DELETE FROM chunks WHERE embedding::text LIKE '%2025-08-%';
