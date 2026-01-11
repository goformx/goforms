<script setup lang="ts">
import { ref, computed } from "vue";
import { useForm } from "@inertiajs/vue3";
import DashboardLayout from "@/components/layout/DashboardLayout.vue";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { CheckCircle2, AlertCircle, Trash2 } from "lucide-vue-next";

interface Props {
  settings?: {
    defaultFormStatus: string;
    notificationsEnabled: boolean;
  };
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

const form = useForm({
  defaultFormStatus: props.settings?.defaultFormStatus ?? "draft",
  notificationsEnabled: props.settings?.notificationsEnabled ?? true,
});

const successMessage = ref<string | null>(null);
const errorMessage = ref<string | null>(null);

const isSubmitting = computed(() => form.processing);

function handleSubmit() {
  successMessage.value = null;
  errorMessage.value = null;

  form.put("/settings", {
    onSuccess: () => {
      successMessage.value = "Settings saved successfully";
    },
    onError: () => {
      errorMessage.value = "Failed to save settings";
    },
  });
}

function handleDeleteAccount() {
  if (!confirm("Are you sure you want to delete your account? This action cannot be undone.")) {
    return;
  }

  form.delete("/account", {
    onError: () => {
      errorMessage.value = "Failed to delete account";
    },
  });
}
</script>

<template>
  <DashboardLayout title="Settings" subtitle="Manage your application settings">
    <div class="space-y-6">
      <!-- Success/Error Messages -->
      <Alert v-if="successMessage || props.flash?.success" variant="success">
        <CheckCircle2 class="h-4 w-4" />
        <AlertDescription>{{ successMessage || props.flash?.success }}</AlertDescription>
      </Alert>

      <Alert v-if="errorMessage || props.flash?.error" variant="destructive">
        <AlertCircle class="h-4 w-4" />
        <AlertDescription>{{ errorMessage || props.flash?.error }}</AlertDescription>
      </Alert>

      <!-- General Settings -->
      <Card class="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle>General Settings</CardTitle>
          <CardDescription>
            Configure your default application settings.
          </CardDescription>
        </CardHeader>
        <form @submit.prevent="handleSubmit">
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label for="defaultFormStatus">Default Form Status</Label>
              <select id="defaultFormStatus" v-model="form.defaultFormStatus"
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2">
                <option value="draft">Draft</option>
                <option value="published">Published</option>
              </select>
              <p class="text-sm text-muted-foreground">
                The default status for newly created forms.
              </p>
            </div>

            <div class="flex items-center space-x-2">
              <input id="notificationsEnabled" v-model="form.notificationsEnabled" type="checkbox"
                class="h-4 w-4 rounded border-input" />
              <Label for="notificationsEnabled">Enable email notifications</Label>
            </div>
          </CardContent>
          <CardFooter>
            <Button type="submit" :disabled="isSubmitting">
              <span v-if="isSubmitting">Saving...</span>
              <span v-else>Save Settings</span>
            </Button>
          </CardFooter>
        </form>
      </Card>

      <!-- Danger Zone -->
      <Card class="bg-card/50 backdrop-blur-sm border-destructive/50">
        <CardHeader>
          <CardTitle class="text-destructive">Danger Zone</CardTitle>
          <CardDescription>
            Irreversible and destructive actions.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div class="flex items-center justify-between">
            <div>
              <h4 class="font-medium">Delete Account</h4>
              <p class="text-sm text-muted-foreground">
                Permanently delete your account and all associated data.
              </p>
            </div>
            <Button variant="destructive" @click="handleDeleteAccount">
              <Trash2 class="mr-2 h-4 w-4" />
              Delete Account
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </DashboardLayout>
</template>
