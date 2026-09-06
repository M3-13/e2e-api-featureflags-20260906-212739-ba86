VERDICT: CHANGES_REQUESTED

## Prüfbericht – Feature-Flag-Service (Go-Backend)

### 1. GDPR (DSGVO)

#### Befund 1: Fehlende Transportverschlüsselung (TLS)
- **Schweregrad:** hoch
- **Beschreibung:** `main.go` startet den Server mit `http.ListenAndServe(":"+port, handler)` ohne TLS. Die Route `GET /flags/{key}/evaluate?user=...` überträgt den `user`-Parameter (personenbezogene Daten) im Klartext, sofern kein TLS-terminierender Reverse-Proxy vorgeschaltet ist. Da der Server an alle Interfaces bindet (`":8080"`), ist eine unverschlüsselte Erreichbarkeit wahrscheinlich. Dies verletzt Art. 32 DSGVO (Sicherheit der Verarbeitung) und die CRA-Anforderungen an sichere Kommunikation.
- **Konkrete Behebung:** In `main.go` TLS aktivieren, z. B. `http.ListenAndServeTLS(":"+port, certFile, keyFile, handler)`. Alternativ im README eindeutig dokumentieren, dass der Dienst ausschließlich hinter einem TLS-terminierenden Reverse-Proxy (nginx, Load Balancer) betrieben werden darf, und diese Vorgabe im Deployment durchsetzen.

#### Befund 2: Fehlende Authentifizierung und Zugriffskontrolle
- **Schweregrad:** hoch
- **Beschreibung:** Die REST-API hat keinerlei Authentifizierung oder Autorisierung. Jeder mit Netzwerkzugriff kann Flags anlegen, ändern, löschen und evaluieren. Dies verletzt Art. 25 DSGVO (Datenschutz durch Technikgestaltung, Zugriffskontrolle) und ist zugleich ein CRA-Verstoß (security by design). Unautorisierte Änderungen können die Produktfunktionalität beeinträchtigen und die Einschleusung personenbezogener Daten über Flag-Beschreibungen ermöglichen.
- **Konkrete Behebung:** In `main.go` eine Authentifizierungs-/Autorisierungs-Middleware einfügen (z. B. API-Key, mTLS, OAuth2). Schreibzugriffe (POST, PUT, DELETE) nur für autorisierte Clients zulassen; Lesezugriffe können offen bleiben oder ebenfalls gesichert werden. Falls der Dienst nur in einem privaten, vertrauenswürdigen Netz betrieben wird, dies im README klar festhalten und die Netzwerksegmentierung dokumentieren.

#### Befund 3: Potenzielle Protokollierung personenbezogener Daten über Flag-Keys
- **Schweregrad:** mittel
- **Beschreibung:** `internal/middleware/logging.go` loggt `r.URL.Path`, das den Flag-Key enthält. Der Key wird in `internal/handlers/crud.go` nur auf leer geprüft, nicht auf erlaubte Zeichen oder Länge. Ein Betreiber oder Angreifer könnte daher Keys mit personenbezogenen Daten (z. B. E-Mail-Adresse) anlegen; diese erscheinen dann in den Logs. Das verletzt den Grundsatz der Datenminimierung (Art. 5 Abs. 1 lit. c DSGVO).
- **Konkrete Behebung:** In `internal/handlers/crud.go` eine Key-Validierung ergänzen: nur `[a-zA-Z0-9._-]{1,128}` zulassen. Alternativ im Logging nur den normalisierten Endpunkt ohne Schlüssel loggen (z. B. per Regex maskieren) oder eine Positivliste von Endpunkten verwenden.

#### Befund 4: Fehlende Datenschutzdokumentation / Rechtsgrundlage
- **Schweregrad:** mittel
- **Beschreibung:** Es gibt keine Datenschutzerklärung, kein Verarbeitungsverzeichnis und keinen Hinweis auf die Rechtsgrundlage für die kurzzeitige Verarbeitung des `user`-Parameters. Der Betreiber muss die Verarbeitung dokumentieren und betroffene Personen informieren (Art. 13, 30 DSGVO), auch wenn der Dienst kein Endnutzer-UI besitzt.
- **Konkrete Behebung:** Im `README.md` oder in einer separaten `PRIVACY.md` die Verarbeitung beschreiben: Zweck (deterministische Feature-Evaluierung), Datenkategorien (Nutzerkennung), Rechtsgrundlage (z. B. Art. 6 Abs. 1 lit. b oder f DSGVO), Speicherdauer (keine dauerhafte Speicherung, nur flüchtige Verarbeitung im Arbeitsspeicher), Empfänger (keine), Betroffenenrechte. Zusätzlich einen Auftragsverarbeitungsvertrag (AVV) bereitstellen, falls der Dienst im Auftrag Dritter betrieben wird.

### 2. EU Cyber Resilience Act (CRA)

#### Befund 5: Fehlende SBOM und dokumentierte Sicherheitseigenschaften
- **Schweregrad:** mittel
- **Beschreibung:** Es gibt keine SBOM-Datei (z. B. CycloneDX/SPDX) und keine `SECURITY.md`. Die CRA verlangt für Produkte mit digitalen Elementen eine SBOM sowie dokumentierte Sicherheitseigenschaften.
- **Konkrete Behebung:** Eine SBOM erstellen (z. B. mit `syft` oder `govulncheck`) und als `sbom.json` im Repository einchecken. Eine `SECURITY.md` mit Sicherheitskontakt, Meldeverfahren für Schwachstellen, Support-Zeitraum und Update-Verpflichtung anlegen und im `README.md` darauf verweisen.

#### Befund 6: Kein dokumentierter Update-/Patch-Prozess
- **Schweregrad:** niedrig
- **Beschreibung:** Es ist kein Prozess erkennbar, wie Sicherheitsupdates bereitgestellt werden. Die CRA verpflichtet Hersteller, während des Support-Zeitraums Sicherheitsupdates bereitzustellen.
- **Konkrete Behebung:** Im `README.md` oder in der `SECURITY.md` einen Abschnitt zum Update-/Patch-Verfahren aufnehmen. Eine Versionsnummerierung (SemVer) einführen und im Health-Endpoint (`/healthz`) oder über einen neuen `/version`-Endpunkt die aktuelle Version ausgeben, damit Updates nachvollziehbar sind.

### 3. EU AI Act
Nicht anwendbar: Das Produkt enthält keine KI-Funktionen.

### 4. Pflichttexte & UI
Nicht anwendbar: Reines Backend ohne Endnutzer-UI; daher keine Legal-Notice-, Cookie-, Widerrufs- oder Barrierefreiheitspflichten.

### 5. Barrierefreiheit
Nicht anwendbar: Kein öffentliches Web-UI; die EAA/WCAG/BITV-Anforderungen greifen nicht.

## Fazit
Die Kernfunktionalität ist solide umgesetzt: Datenminimierung beim Nutzer-Parameter, Body-Limit, saubere Fehlerantworten und Race-freie Speicherung sind vorhanden. Es bestehen jedoch wesentliche Sicherheits- und Dokumentationslücken (TLS, Zugriffskontrolle, SBOM/Privacy-Doku), die vor einem Markteinsatz behoben werden müssen. Nach Umsetzung der genannten Maßnahmen kann die Freigabe erfolgen.