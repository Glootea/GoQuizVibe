package com.glootea.goquiz

import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.compositionLocalOf
import com.glootea.goquiz.core.di.AppGraph
import com.glootea.goquiz.navigation.AppNavRoot

val LocalAppGraph = compositionLocalOf<AppGraph> {
    error("AppGraph not provided. Wrap your composable in AppContent(graph).")
}

@Composable
fun AppContent(graph: AppGraph) {
    CompositionLocalProvider(LocalAppGraph provides graph) {
        AppNavRoot()
    }
}
