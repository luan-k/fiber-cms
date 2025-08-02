import { useState } from "react";

export default function PostsReloader({ initialPosts, totalPosts }) {
  const [posts, setPosts] = useState(initialPosts);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [lastUpdated, setLastUpdated] = useState(new Date());

  const reloadPosts = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/posts-reload");
      if (!response.ok) {
        throw new Error(`Failed to fetch: ${response.status}`);
      }

      const data = await response.json();
      setPosts(data.slice(0, 3));
      setLastUpdated(new Date());
    } catch (err) {
      setError(err.message);
      console.error("Error reloading posts:", err);
    } finally {
      setLoading(false);
    }
  };

  const formatTime = (date) => {
    return date.toLocaleTimeString("en-US", {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  };

  return (
    <div className='posts-reloader'>
      <div className='reload-header'>
        <h2>Latest Posts</h2>
        <div className='reload-controls'>
          <span className='last-updated'>
            Last updated: {formatTime(lastUpdated)}
          </span>
          <button
            onClick={reloadPosts}
            disabled={loading}
            className={`reload-btn ${loading ? "loading" : ""}`}>
            {loading ? (
              <>
                <span className='spinner'></span>
                Reloading...
              </>
            ) : (
              <>🔄 Reload Posts</>
            )}
          </button>
        </div>
      </div>

      {error && (
        <div className='reload-error'>
          <strong>Reload Error:</strong> {error}
        </div>
      )}

      <div className='posts-grid'>
        {posts.map((post) => (
          <article key={post.id} className='post-card'>
            <h3 className='post-title'>{post.title}</h3>
            <div className='post-meta'>
              By {post.username} •{" "}
              {new Date(post.created_at).toLocaleDateString()}
            </div>
            <p className='post-description'>{post.description}</p>
            <a href={`/posts/${post.id}`} className='read-more'>
              Read more →
            </a>
          </article>
        ))}
      </div>

      <div className='view-all'>
        <a href='/posts'>View All {totalPosts} Posts</a>
      </div>

      <style jsx>{`
        .posts-reloader {
          margin-top: 4rem;
        }

        .reload-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 2rem;
          flex-wrap: wrap;
          gap: 1rem;
        }

        .reload-header h2 {
          margin: 0;
        }

        .reload-controls {
          display: flex;
          align-items: center;
          gap: 1rem;
          flex-wrap: wrap;
        }

        .last-updated {
          font-size: 0.9rem;
          color: #666;
          font-style: italic;
        }

        .reload-btn {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.5rem 1rem;
          background: var(--accent, #8b5cf6);
          color: white;
          border: none;
          border-radius: 8px;
          cursor: pointer;
          font-weight: 500;
          transition: all 0.2s ease;
          min-width: 140px;
          justify-content: center;
        }

        .reload-btn:hover:not(:disabled) {
          background: var(--accent-dark, #7c3aed);
          transform: translateY(-1px);
        }

        .reload-btn:disabled {
          opacity: 0.7;
          cursor: not-allowed;
          transform: none;
        }

        .spinner {
          width: 16px;
          height: 16px;
          border: 2px solid transparent;
          border-top: 2px solid currentColor;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          to {
            transform: rotate(360deg);
          }
        }

        .reload-error {
          background: #fee;
          border: 1px solid #fcc;
          color: #c00;
          padding: 1rem;
          border-radius: 8px;
          margin-bottom: 2rem;
        }

        .posts-grid {
          display: grid;
          gap: 2rem;
          margin-bottom: 2rem;
        }

        .post-card {
          border: 1px solid #e5e7eb;
          border-radius: 12px;
          padding: 1.5rem;
          background: white;
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
          transition: transform 0.2s ease, box-shadow 0.2s ease;
        }

        .post-card:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        }

        .post-title {
          color: #1f2937;
          font-size: 1.25rem;
          margin: 0 0 0.5rem 0;
          line-height: 1.2;
        }

        .post-meta {
          color: #6b7280;
          font-size: 0.9rem;
          margin-bottom: 1rem;
        }

        .post-description {
          color: #374151;
          line-height: 1.6;
          margin-bottom: 1rem;
        }

        .read-more {
          color: var(--accent, #8b5cf6);
          text-decoration: none;
          font-weight: 500;
        }

        .read-more:hover {
          text-decoration: underline;
        }

        .view-all {
          text-align: center;
        }

        .view-all a {
          color: var(--accent, #8b5cf6);
          text-decoration: none;
          font-weight: 600;
          padding: 0.75rem 2rem;
          border: 2px solid var(--accent, #8b5cf6);
          border-radius: 8px;
          transition: all 0.2s ease;
          display: inline-block;
        }

        .view-all a:hover {
          background: var(--accent, #8b5cf6);
          color: white;
        }

        @media (max-width: 640px) {
          .reload-header {
            flex-direction: column;
            align-items: flex-start;
          }

          .reload-controls {
            width: 100%;
            justify-content: space-between;
          }
        }
      `}</style>
    </div>
  );
}
