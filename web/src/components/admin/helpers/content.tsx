import React from "react";
import Table, { type TableColumnWithRender } from "@/components/admin/ui/Table";
import PostTitle from "@/components/admin/ui/PostTitle";
import { formatDateTime } from "@/utils/formatting";
import type { Post } from "@/lib/types";

interface AdminContentTableProps {
  recentPosts: Post[];
}

const columns: TableColumnWithRender<Post>[] = [
  { key: "title", name: "Title", width: "34.8125rem", render: (_, row) => <PostTitle value={row} /> },
  { key: "created_at", name: "Date", width: "10.75rem", render: (value) => formatDateTime(String(value)) },
];

const AdminContent: React.FC<AdminContentTableProps> = ({ recentPosts }) => (
  <Table
    columns={columns}
    data={recentPosts}
  />
);

export default AdminContent;