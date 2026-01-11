<script setup lang="ts">
import { computed } from "vue";
import { Link } from "@inertiajs/vue3";
import DashboardLayout from "@/components/layout/DashboardLayout.vue";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Card, CardContent } from "@/components/ui/card";
import { Pencil, Eye, Download } from "lucide-vue-next";

interface Submission {
  id: string;
  data: Record<string, unknown>;
  status: string;
  createdAt: string;
  updatedAt: string;
}

interface Form {
  id: string;
  title: string;
}

interface Props {
  form: Form;
  submissions: Submission[];
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

const hasSubmissions = computed(() => props.submissions && props.submissions.length > 0);

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function getStatusVariant(status: string): "default" | "success" | "warning" {
  switch (status) {
    case "completed":
      return "success";
    case "pending":
      return "warning";
    default:
      return "default";
  }
}

function getSubmissionPreview(data: Record<string, unknown>): string {
  const values = Object.values(data).slice(0, 3);
  const preview = values
    .filter((v) => typeof v === "string" || typeof v === "number")
    .join(", ");
  return preview.length > 50 ? `${preview.substring(0, 50)}...` : preview || "No data";
}

function exportSubmissions() {
  // Create CSV from submissions
  const headers = ["ID", "Status", "Created At", "Data"];
  const rows = props.submissions.map((s) => [
    s.id,
    s.status,
    s.createdAt,
    JSON.stringify(s.data),
  ]);

  const csv = [
    headers.join(","),
    ...rows.map((row) => row.map((cell) => `"${cell}"`).join(",")),
  ].join("\n");

  const blob = new Blob([csv], { type: "text/csv" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${props.form.title}-submissions.csv`;
  a.click();
  URL.revokeObjectURL(url);
}
</script>

<template>
  <DashboardLayout :title="`Submissions: ${props.form.title}`" subtitle="View and manage form submissions">
    <template #actions>
      <Button variant="outline" as-child>
        <Link :href="`/forms/${props.form.id}/edit`">
          <Pencil class="mr-2 h-4 w-4" />
          Edit Form
        </Link>
      </Button>
      <Button v-if="hasSubmissions" variant="outline" @click="exportSubmissions">
        <Download class="mr-2 h-4 w-4" />
        Export CSV
      </Button>
    </template>

    <!-- Empty State -->
    <Card v-if="!hasSubmissions" class="bg-card/50 backdrop-blur-sm border-border/50">
      <CardContent class="flex flex-col items-center justify-center py-12">
        <div class="text-center">
          <h3 class="text-lg font-semibold">No submissions yet</h3>
          <p class="text-muted-foreground mt-1">
            This form hasn't received any submissions yet.
          </p>
          <Button class="mt-4" variant="outline" as-child>
            <Link :href="`/forms/${props.form.id}/preview`">
              <Eye class="mr-2 h-4 w-4" />
              Preview Form
            </Link>
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Submissions Table -->
    <Card v-else class="bg-card/50 backdrop-blur-sm border-border/50">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>ID</TableHead>
            <TableHead>Preview</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Submitted</TableHead>
            <TableHead class="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="submission in submissions" :key="submission.id">
            <TableCell class="font-mono text-sm">
              {{ submission.id.substring(0, 8) }}...
            </TableCell>
            <TableCell class="max-w-[300px] truncate text-muted-foreground">
              {{ getSubmissionPreview(submission.data) }}
            </TableCell>
            <TableCell>
              <Badge :variant="getStatusVariant(submission.status)">
                {{ submission.status }}
              </Badge>
            </TableCell>
            <TableCell class="text-muted-foreground">
              {{ formatDate(submission.createdAt) }}
            </TableCell>
            <TableCell class="text-right">
              <Button variant="ghost" size="sm" title="View Details">
                <Eye class="h-4 w-4" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </Card>
  </DashboardLayout>
</template>
