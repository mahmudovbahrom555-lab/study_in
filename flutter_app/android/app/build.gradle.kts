import java.net.URI
import java.util.Base64
import java.util.Properties

plugins {
    id("com.android.application")
    id("kotlin-android")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

val releaseSigningFile = rootProject.file("key.properties")
val releaseSigning = Properties().apply {
    if (releaseSigningFile.isFile) {
        releaseSigningFile.inputStream().use { load(it) }
    }
}

android {
    namespace = "uz.mamulearn.app"
    compileSdk = maxOf(36, flutter.compileSdkVersion)
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = JavaVersion.VERSION_17.toString()
    }

    defaultConfig {
        // Не менять после публикации в Google Play (DECISIONS.md, 2026-10-06).
        applicationId = "uz.mamulearn.app"
        // You can update the following values to match your application needs.
        // For more information, see: https://flutter.dev/to/review-gradle-config.
        minSdk = flutter.minSdkVersion
        targetSdk = maxOf(36, flutter.targetSdkVersion)
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    signingConfigs {
        if (releaseSigningFile.isFile) {
            create("release") {
                fun signingProperty(name: String): String =
                    requireNotNull(releaseSigning.getProperty(name)?.takeIf { it.isNotBlank() }) {
                        "Set $name in android/key.properties"
                    }

                storeFile = rootProject.file(signingProperty("storeFile"))
                storePassword = signingProperty("storePassword")
                keyAlias = signingProperty("keyAlias")
                keyPassword = signingProperty("keyPassword")
            }
        }
    }

    buildTypes {
        release {
            signingConfig = signingConfigs.findByName("release")
        }
    }
}

val releaseDartDefines = providers.gradleProperty("dart-defines").orElse("")

val validateReleaseConfiguration by tasks.registering {
    inputs.property("dartDefines", releaseDartDefines)
    doLast {
        val defines = releaseDartDefines.get().split(',').filter { it.isNotEmpty() }
            .map { Base64.getDecoder().decode(it).toString(Charsets.UTF_8) }
        val environments = defines.filter { it.startsWith("ENV=") }
        check(environments == listOf("ENV=production")) {
            "Set --dart-define=ENV=production once for the Android release."
        }
        val apiUrls = defines.filter { it.startsWith("API_URL=") }
        val apiUri = apiUrls.singleOrNull()?.substringAfter('=')?.let {
            runCatching { URI(it) }.getOrNull()
        }
        check(apiUri != null && apiUri.scheme == "https" &&
                apiUri.host?.contains('.') == true &&
                apiUri.host !in setOf("10.0.2.2", "127.0.0.1") &&
                apiUri.userInfo == null && apiUri.query == null && apiUri.fragment == null) {
            "Set --dart-define=API_URL to the production HTTPS API URL once for the Android release."
        }
        check(releaseSigningFile.isFile) {
            "Create android/key.properties with the Google Play upload key before building release."
        }
    }
}

tasks.matching { it.name == "preReleaseBuild" || it.name == "compileFlutterBuildRelease" }.configureEach {
    dependsOn(validateReleaseConfiguration)
}

flutter {
    source = "../.."
}
