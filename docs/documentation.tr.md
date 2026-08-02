Kivi

fişlerini gösteren veritabanı · v1.1.0 · biçim v2

Saf Go ile yazılmış bir olay-defteri veritabanı — verdiği her cevabın hesabını verebilen tek statik ikili.

## Kivi nedir

Kivi yalnızca **olayları** saklar — değişmez, özet-zincirli kayıtlar. Güncel duruma benzeyen her şey (bir tablo, bir graf, bir zaman serisi, bir vektör indeksi) bir **görünümdür**: bu olayların bir katlaması, talep üzerine yeniden derlenen ve her an atılabilir. Her cevap bir **iz** (onu kuran kayıt numaraları) ve bir **kapsam** (geçmişe ne kadar baktığı) taşır ve `why` fiş kayıtlarının kendisini döndürür. Eksik veri dürüst bir rettir, asla bir uydurma değil.

## Üç ilke

**Kalıcı olan yalnız olaylardır.**  
UPDATE yoktur. Yeni olgular eklenir; geçmiş asla yeniden yazılmaz. Tarihçe birinci sınıftır.

**Temsiller derlenir.**  
`derived/`'i silmek her zaman meşrudur: cevaplar değişmez, yalnız bir sonraki sorgu yavaşlar.

**İzsiz cevap yok.**  
İzsiz bir cevap temsil edilemez — çekirdekte ve her SDK'da. Hiçbir şey uydurulmaz.

## Normal bir veritabanından farkı

|  | Tipik bir veritabanı | Kivi |
| --- | --- | --- |
| Kalıcı olan | Güncel durum; UPDATE geçmişi yok eder | Olaylar; durum onlardan yeniden derlenir |
| Cevaplar | Çıplak değerler | Değer + iz + kapsam; `why` fişleri getirir |
| Zaman-noktası | Fotoğraflar ve ek mekanizma | Tasarımı gereği bedava: N kaydına kadar katla (`--as-of`) |
| Bütünlük | Varsayılır | Doğrulanır: SHA-256 zinciri + Ed25519 mühürleri; istemciler sunucuyu yeniden doğrular |
| Silme | Satır gitti, hikâye gitti | Kripto-silme: anahtar yok edilir, baytlar değişmez, silme fişlenir |

## Hızlı başlangıç

Sunucuyu çalıştır (tek statik ikili ya da konteyner imajı):

```
# Docker
docker run -p 4741:4741 -p 4742:4742 -e KIVI_TOKEN=change-me iwasoftcom/kivi:1.1.0

# ya da ikili
KIVI_DATA=/var/lib/kivi/kivi.ledger KIVI_TOKEN=change-me kivid
```

Bir olay ekle ve geri oku — cevap iziyle birlikte gelir:

```
kivi append ./kivi.ledger property '{"subject":"dog","attribute":"sound","value":"woof"}'
kivi table  ./kivi.ledger dog sound
# → {"scope":0,"trace":[0],"value":"woof"}

kivi verify ./kivi.ledger        # zinciri yeniden özetle, her mührü doğrula
kivi why    ./kivi.ledger 0      # cevabın arkasındaki fiş
```

İstemciler aynı sözleşmeyi **altı dilde** konuşur — Go, Python, Java & Kotlin, Rust, Node.js ve .NET — tipli "entity ver, entity al" erişimi ve varsayılan açık istemci-tarafı doğrulamayla.

## İçinde ne var

**İzli cevaplar**  
Her değer, onu kuran olayları gösterir; izsiz bir cevap oluşturulamaz.

**Zaman yolculuğu**  
"N kaydında ne biliyorduk?" tek bir sınırlı replay — fotoğraf yok, migration yok (`--as-of`).

**Doğrulanmış bütünlük**  
SHA-256 özet zinciri + Ed25519 mühürleri; tek bir bozuk bayt yakalanır ve adıyla söylenir.

**Kripto-silme**  
Kayıt başına anahtar; silmek anahtarı yok eder, zinciri geçerli tutar ve silmenin kendisini fişler.

**ACID, ekleme-tabanlı biçim**  
Olay başına atomiklik, sıralı yazmalar, snapshot okumalar, grup-commit `fsync` dayanıklılığı.

**Küme & federasyon**  
Çoğunluk onaylı failover, okuma kopyaları ve bağımsız defterler arasında karşılıklı tanıklık.

**Yönetim paneli & kimlik**  
Gömülü UI, fişli yapılandırma, kullanıcı/rol/oturum ve adlı, iptal edilebilir API anahtarları.

**LLM kapısı (MCP)**  
Bir yapay-zekâ ajanına kaynağını gösteren bir bellek verir; eksik bir olgu uydurulmaz, reddedilir.

**Çok kiracılı, talep üzerine**  
Tek süreçte birbirinden yalıtılmış birden çok kiracı — ve yenisi çalışma anında, admin API üzerinden, fişlenerek açılabilir; yeniden başlatma yok.

**Servis olarak çalışır**  
Linux'ta systemd (.deb/.rpm), Windows'ta installer'dan gerçek bir Windows servisi — açılışta başlar, dururken düzgün boşaltır.

## Yönetim paneli

Her sunucu, ops portunda gömülü bir web paneli getirir (ayrı kurulum yok). `https://<host>:4742/admin/` adresini aç ve bir kullanıcı hesabıyla giriş yap — oradan, canlı ve veri düzlemine dokunmadan şunları yapabilirsin:

-   **Genel bakış & Dönemler** — sağlık, kayıt ve mühür sayıları, rotasyon.
-   **Kayıtlar & Arama** — eklemelerin canlı akışı, artı anlamsal arama.
-   **Küme & Federasyon** — topoloji, failover ve tanıklık durumu.
-   **Kullanıcılar & API anahtarları** — hesap ve rol oluştur, anahtar ver/iptal et.
-   **Config** — limitleri ve ayarları yeniden ayarla; her değişiklik deftere fişlenir (sırlar asla fişlenmez).

## Mimari

Tek süreç, iki ağ kapısı, katı bir bağımlılık kuralı: çekirdek yalnız Go standart kütüphanesini kullanır; gRPC kenarda yaşar. Diskte tam olarak tek tür kalıcı gerçek vardır — olay defteri — ve geri kalan her şey atılabilir.

CLI · altı-dilli SDK'lar güvenmeyen: istemci-tarafı doğrular Yönetim paneli · LLM ajanları (MCP) gömülü UI · kaynak-gösteren bellek gRPC veri düzlemi · :4741 ops HTTP · :4742 çekirdek (Go stdlib): defter · görünümler · iz · mühür · kimlik cevaplar sorgu anında derlenir — türetilen hiçbir şey birincil değil disk: yalnızca-ekleme olay defteri (+ mühürler)

## İstemci SDK’ları

Altı istemci SDK’sı aynı tel sözleşmesini konuşur — hepsi MIT lisanslı ve her dilin registry’sine yayınlanmıştır. kivi sunucusu ve çekirdeği tescilli kalır; istemciler açıktır.

| Dil | Registry | Kurulum | Paket sayfası |
| --- | --- | --- | --- |
| Rust | crates.io | `cargo add kivi-sdk` | [crates.io/crates/kivi-sdk](https://crates.io/crates/kivi-sdk) |
| Python | PyPI | `pip install kivi-sdk` | [pypi.org/project/kivi-sdk](https://pypi.org/project/kivi-sdk/) |
| Node.js | npm | `npm install @iwasoft/kivi` | [npmjs.com/package/@iwasoft/kivi](https://www.npmjs.com/package/@iwasoft/kivi) |
| Java / Kotlin | Maven Central | `com.iwasoft:kivi:1.1.0` | [central.sonatype.com/artifact/com.iwasoft/kivi](https://central.sonatype.com/artifact/com.iwasoft/kivi) |
| .NET | NuGet | `dotnet add package Iwasoft.Kivi` | [nuget.org/packages/Iwasoft.Kivi](https://www.nuget.org/packages/Iwasoft.Kivi) |
| Go | GitHub | `go get github.com/iwasoftcom/kivi-sdk` | [github.com/iwasoftcom/kivi-sdk](https://github.com/iwasoftcom/kivi-sdk) |

### Güvenmeyen istemci

Her SDK aynı küçük, tipli yüzeyi sunar — ve istemci tarafı doğrulama **varsayılan olarak açıktır**: istemci okurken zinciri yeniden özetler ve Ed25519 mühürlerini denetler, böylece yalan söyleyen bir sunucu ya da tek bir ters çevrilmiş bayt güvenilmeden yakalanır.

-   `append(type, body)` — bir olay ekle, fişini al.
-   `table(subject, attribute)` — **izli bir cevap**: `value` + `trace` (onu kuran kayıt numaraları) + `scope`. Eksik bir hücre dürüst bir rettir — yerel bir istisna, asla uydurulmuş bir null değil.
-   `table(…, as_of=N)` — aynı cevap **N kaydı anına** (zaman yolculuğu).
-   `why(trace)` — fiş kayıtlarının kendisi.
-   `replay()` — her kaydı akıt, istemci tarafında yeniden doğrulanmış.
-   `similar(query, k)` — izli anlamsal arama (kayıt + skor).
-   `login(user, pw)` — role-kapsamlı bir oturum belirteci; `with_bearer` kanalı başka bir kimlik altında yeniden kullanır.
-   `head()` — ucuz yönelim (head numarası + özet), denetimsiz.

### Örnek (Python)

Biçim altı dilin hepsinde birebir aynıdır:

```
from kivi import KiviClient

c = KiviClient("localhost:4741", token="…")     # verify=True by default
c.append("property", {"subject": "dog", "attribute": "sound", "value": "bark"})
a = c.table("dog", "sound")              # TracedAnswer(value="bark", trace=[0], scope=0)
old = c.table("dog", "sound", as_of=41)  # the answer as of record 41
receipts = c.why(a.trace)                # the actual ledger records
for rec in c.replay():                   # hash + chain + seal verified CLIENT-SIDE
    ...
```

Tek bir konformans sınavı altısına birden koşar, böylece her dil aynı cevabı aynı izle döndürür. Dil bazlı örnekler (Go, Java / Kotlin, Rust, Node.js, .NET) ve tipli entity katmanı [tam referansta](reference.en.html) yer alır.

## Tam dokümantasyon & kaynak

-   **[Tam referans (İngilizce)](reference.en.html)** — her özellik, CLI ve gRPC API'si, ortam değişkenleri, nasıl-yapılır kılavuzları ve dil-sekmeli kod örnekleri.
-   **Uyumluluk:** gRPC API'si, disk biçimi v2 ve SDK yüzeyi bir SemVer sözleşmesidir — major sürüm artışı olmadan kırılmazlar.
-   **Dürüst durum:** henüz bağımsız güvenlik denetiminden geçmedi; henüz üretim kilometresi yok. Bunlar birer ifşadır, kararlılık sözüne çekince değil.

## İletişim

Sorular, demo ya da tasarım-ortağı görüşmesi için bize ulaşın:

-   **E-posta** — [info@iwasoft.com](mailto:info@iwasoft.com)
-   **LinkedIn** — [linkedin.com/company/iwasoft](https://www.linkedin.com/company/iwasoft)

[İletişim](#contact) · Kivi v1.1.0 · disk biçimi v2 (donmuş sözleşme, altın vektörler) · Go stdlib çekirdek, kenarlarda gRPC · tek statik ikili. © iwasoft.
