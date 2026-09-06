# Sicherheit

Dieses Dokument beschreibt den Umgang mit Sicherheitsvorfällen, den
Support-Zeitraum und die verbindlichen Betriebsvorgaben für
`e2e-api-featureflags`.

## Sicherheitskontakt

Sicherheitsrelevante Meldungen richten Sie bitte per E-Mail an:

**security@example.com**

Bitte verwenden Sie diesen Kanal ausschließlich für Sicherheitsmeldungen.
Geben Sie dabei eine Beschreibung des Problems, betroffene Komponenten und –
sofern verfügbar – Schritte zur Reproduktion an.

## Meldeverfahren für Schwachstellen

1. **Meldung:** Senden Sie eine E-Mail an den oben genannten Sicherheitskontakt.
   Bitte veröffentlichen Sie die Schwachstelle nicht, bevor sie behoben wurde
   (verantwortungsvolle Offenlegung, „Responsible Disclosure").
2. **Bestätigung:** Sie erhalten innerhalb von 5 Werktagen eine
   Eingangsbestätigung.
3. **Bewertung:** Die Schwachstelle wird bewertet und priorisiert.
4. **Behebung:** Es wird ein Fix erstellt und als neue Version bereitgestellt
   (siehe „Update-/Patch-Verfahren").
5. **Veröffentlichung:** Nach der Behebung wird eine Sicherheitshinweis
   veröffentlicht und der Meldende – auf Wunsch – namentlich genannt.

## Support-Zeitraum

| Version        | Support             |
| -------------- | ------------------- |
| Aktuelle Major | Sicherheits- und Bugfix-Updates |
| Vorherige Major | Sicherheitsupdates für 6 Monate nach Veröffentlichung der Nachfolgeversion |

Außerhalb des Support-Zeitraums werden keine Sicherheitsupdates mehr
bereitgestellt; ein Betrieb ist dann nicht mehr empfohlen.

## Update-/Patch-Verfahren

- Der Dienst folgt **Semantic Versioning (SemVer)**: `MAJOR.MINOR.PATCH`.
  Die aktuelle Version ist im Health-Endpoint (`GET /healthz`) abrufbar.
- Sicherheitsfixes werden als `PATCH`-Release (bzw. bei notwendigen
  Breaking-Changes als `MAJOR`-Release) veröffentlicht.
- Betreiber sollten die ausgelieferte Version regelmäßig mit den veröffentlichten
  Versionen abgleichen und Sicherheitsupdates zeitnah einspielen.

## Verbindliche Betriebsvorgaben

- **TLS-Terminierung:** Der Dienst bindet standardmäßig unverschlüsselt an
  alle Interfaces und terminiert selbst kein TLS. Er darf **ausschließlich
  hinter einem TLS-terminierenden Reverse-Proxy** (z. B. nginx, Load Balancer)
  betrieben werden. Eine direkte unverschlüsselte Erreichbarkeit aus
  untrusted Netzen ist untersagt.
- **Zugriffskontrolle:** Der Dienst ist für den Betrieb in vertrauenswürdigen,
  abgeschotteten Netzen vorgesehen. Schreibzugriffe (`POST`, `PUT`, `DELETE`)
  sind durch die vorgeschaltete Infrastruktur (Netzsegmentierung,
  vorgeschaltete Authentifizierung) zu schützen.

## Keine Sicherheitsentscheidungen

Dieser Dienst **darf keine Sicherheitsentscheidungen tragen**. Insbesondere gilt:

- Der clientseitig übergebene `user`-Parameter ist **kein Nachweis der
  Identität** – er wird vom Aufrufer frei gewählt und nicht authentifiziert.
- Für sicherheitsrelevante Freischaltungen muss die Nutzeridentität aus einer
  vorgeschalteten Authentifizierungsschicht bezogen werden, niemals aus dem
  `user`-Parameter dieses Dienstes.
- Das verwendete Hash-Verfahren (FNV-32a) ist bewusst einfach gewählt und
  dient ausschließlich der deterministischen, gleichmäßigen Verteilung von
  Rollouts – nicht dem kryptographischen Schutz von Entscheidungen.
