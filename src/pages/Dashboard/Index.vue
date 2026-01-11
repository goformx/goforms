<script setup lang="ts">
import { computed } from "vue";
import { Link, router } from "@inertiajs/vue3";
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
import { Plus, Eye, Pencil, ListChecks, Trash2 } from "lucide-vue-next";

interface Form {
  id: string;
  title: string;
  description: string;
  status: "draft" | "published" | "archived";
  createdAt: string;
  updatedAt: string;
}

interface Props {
  forms: Form[];
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

const hasForms = computed(() => props.forms && props.forms.length > 0);

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function getStatusVariant(status: string): "default" | "success" | "secondary" {
  switch (status) {
    case "published":
      return "success";
    case "draft":
      return "secondary";
    case "archived":
      return "default";
    default:
      return "secondary";
  }
}

async function deleteForm(formId: string) {
  if (!confirm("Are you sure you want to delete this form?")) {
    return;
  }

  router.delete(`/forms/${formId}`, {
    preserveScroll: true,
  });
}
</script>

<template>
  <DashboardLayout title="Your Forms" subtitle="Manage and create forms">
    <template #actions>
      <Button as-child>
        <Link href="/forms/new">
          <Plus class="mr-2 h-4 w-4" />
          New Form
        </Link>
      </Button>
    </template>

    <!-- Empty State -->
    <Card v-if="!hasForms">
      <CardContent class="flex flex-col items-center justify-center py-12">
        <div class="text-center">
          <h3 class="text-lg font-semibold">No forms yet</h3>
          <p class="text-muted-foreground mt-1">
            You haven't created any forms yet.
          </p>
          <Button class="mt-4" as-child>
            <Link href="/forms/new">
              <Plus class="mr-2 h-4 w-4" />
              Create Your First Form
            </Link>
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Forms Table -->
    <Card v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Title</TableHead>
            <TableHead>Description</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead>Updated</TableHead>
            <TableHead class="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="form in forms" :key="form.id">
            <TableCell class="font-medium">
              <Link
                :href="`/forms/${form.id}/edit`"
                class="hover:underline text-primary"
              >
                {{ form.title }}
              </Link>
            </TableCell>
            <TableCell class="max-w-[200px] truncate text-muted-foreground">
              {{ form.description || "No description" }}
            </TableCell>
            <TableCell>
              <Badge :variant="getStatusVariant(form.status)">
                {{ form.status }}
              </Badge>
            </TableCell>
            <TableCell class="text-muted-foreground">
              {{ formatDate(form.createdAt) }}
            </TableCell>
            <TableCell class="text-muted-foreground">
              {{ formatDate(form.updatedAt) }}
            </TableCell>
            <TableCell class="text-right">
              <div class="flex items-center justify-end gap-2">
                <Button variant="ghost" size="icon" as-child title="Preview">
                  <Link :href="`/forms/${form.id}/preview`">
                    <Eye class="h-4 w-4" />
                  </Link>
                </Button>
                <Button variant="ghost" size="icon" as-child title="Edit">
                  <Link :href="`/forms/${form.id}/edit`">
                    <Pencil class="h-4 w-4" />
                  </Link>
                </Button>
                <Button variant="ghost" size="icon" as-child title="Submissions">
                  <Link :href="`/forms/${form.id}/submissions`">
                    <ListChecks class="h-4 w-4" />
                  </Link>
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  title="Delete"
                  @click="deleteForm(form.id)"
                >
                  <Trash2 class="h-4 w-4 text-destructive" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </Card>
  </DashboardLayout>
</template>
