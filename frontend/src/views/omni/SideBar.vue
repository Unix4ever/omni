<!--
Copyright (c) 2024 Sidero Labs, Inc.

Use of this software is governed by the Business Source License
included in the LICENSE file.
-->
<template>
  <div>
  <t-sidebar-list :items="items"/>
  <div class="border-naturals-N4 border-t" v-if="machines">
    <div class="px-6 text-xs text-naturals-N8 pt-4">
      Machines
    </div>
    <t-sidebar-list :items="machines"/>
    <div class="px-6 text-xs text-naturals-N8 border-t border-naturals-N4 pt-4">
      Machines Autoprovisioned
    </div>
    <t-sidebar-list :items="automaticMachines"/>
    <div class="px-6 text-xs text-naturals-N8 border-t border-naturals-N4 pt-4">
      Saved Filters
    </div>
  </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RouteLocationRaw, useRoute } from "vue-router";

import TSidebarList from "@/components/SideBar/TSideBarList.vue";
import { canManageBackupStore, canManageUsers, canReadClusters, canReadMachines } from "@/methods/auth";
import { setupBackupStatus } from "@/methods";
import { IconType } from "@/components/common/Icon/TIcon.vue";

const route = useRoute();

const getRoute = (name: string, path: string) => {
  return route.query.cluster ? {
        name: name,
        query: {
          cluster: route.query.cluster,
          namespace: route.query.namespace,
          uid: route.query.uid,
        },
      } : path;
};

const { status: backupStatus } = setupBackupStatus();

const items = computed(() => {
  const result = [{
    name: "Home",
    route: getRoute("Overview", "/omni/"),
    icon: "home" as IconType,
  }];

  if (canReadClusters.value) {
    result.push({
      name: "Clusters",
      route: getRoute("Clusters", "/omni/clusters"),
      icon: "clusters",
    });
  }

  if (canReadMachines.value) {
    result.push({
      name: "Machine Classes",
      route: getRoute("MachineClasses", "/omni/machine-classes"),
      icon: "code-bracket",
    });
  }

  if (canManageUsers.value || (backupStatus.value.configurable && canManageBackupStore.value)) {
    result.push({
      name: "Settings",
      route: getRoute("Settings", "/omni/settings"),
      icon: "settings",
    });
  }

  return result;
});

const machines = computed(() => {
  const result: {name: string, icon: IconType, route: RouteLocationRaw}[] = [];

  if (canReadMachines.value) {
    result.push({
      name: "All Machines",
      route: getRoute("Machines", "/omni/machines"),
      icon: "nodes",
    });

    result.push({
      name: "Self-Managed",
      route: getRoute("MachinesManual", "/omni/machines/manual"),
      icon: "cpu-chip",
    });
  }

  return result;
});

const automaticMachines = computed(() => {
  const result: {name: string, icon: IconType, route: RouteLocationRaw}[] = [];

  if (canReadMachines.value) {
    result.push({
      name: "Virtual",
      route: getRoute("MachinesVirtual", "/omni/machines/infra/virtual"),
      icon: "cloud-connection",
    });

    result.push({
      name: "Physical",
      route: getRoute("MachinesPhysical", "/omni/machines/infra/physical"),
      icon: "server-network",
    });
  }

  return result;
});
</script>
