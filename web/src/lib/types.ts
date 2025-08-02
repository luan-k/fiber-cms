export interface User {
  id: number;
  username: string;
  full_name: string;
  email: string;
  role: string;
  created_at: string;
  password_changed_at?: string;
}

export interface Post {
  id: number;
  title: string;
  description: string;
  content: string;
  user_id: number;
  username: string;
  url: string;
  created_at: string;
  changed_at: string;
}

export interface Taxonomy {
  id: number;
  name: string;
  description: string;
}

export interface Media {
  id: number;
  filename: string;
  original_name: string;
  mime_type: string;
  size: number;
  path: string;
  url: string;
  user_id: number;
  created_at: string;
}
