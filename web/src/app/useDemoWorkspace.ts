import { useReducer } from "react"
import { createDemoWorkspaceActions } from "@/demo/workspace-actions"
import { createDemoWorkspaceState, demoWorkspaceReducer } from "@/demo/workspace-state"

export function useDemoWorkspace(actorId: string) {
  const [state, dispatch] = useReducer(demoWorkspaceReducer, undefined, createDemoWorkspaceState)
  return { state, actions: createDemoWorkspaceActions(state, dispatch, actorId) }
}
