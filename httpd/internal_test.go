package httpd

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/drakkan/sftpgo/common"
	"github.com/drakkan/sftpgo/dataprovider"
	"github.com/drakkan/sftpgo/kms"
	"github.com/drakkan/sftpgo/utils"
	"github.com/drakkan/sftpgo/vfs"
)

const (
	httpdCert = `-----BEGIN CERTIFICATE-----
MIICHTCCAaKgAwIBAgIUHnqw7QnB1Bj9oUsNpdb+ZkFPOxMwCgYIKoZIzj0EAwIw
RTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGElu
dGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMDAyMDQwOTUzMDRaFw0zMDAyMDEw
OTUzMDRaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYD
VQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwdjAQBgcqhkjOPQIBBgUrgQQA
IgNiAARCjRMqJ85rzMC998X5z761nJ+xL3bkmGVqWvrJ51t5OxV0v25NsOgR82CA
NXUgvhVYs7vNFN+jxtb2aj6Xg+/2G/BNxkaFspIVCzgWkxiz7XE4lgUwX44FCXZM
3+JeUbKjUzBRMB0GA1UdDgQWBBRhLw+/o3+Z02MI/d4tmaMui9W16jAfBgNVHSME
GDAWgBRhLw+/o3+Z02MI/d4tmaMui9W16jAPBgNVHRMBAf8EBTADAQH/MAoGCCqG
SM49BAMCA2kAMGYCMQDqLt2lm8mE+tGgtjDmtFgdOcI72HSbRQ74D5rYTzgST1rY
/8wTi5xl8TiFUyLMUsICMQC5ViVxdXbhuG7gX6yEqSkMKZICHpO8hqFwOD/uaFVI
dV4vKmHUzwK/eIx+8Ay3neE=
-----END CERTIFICATE-----`
	httpdKey = `-----BEGIN EC PARAMETERS-----
BgUrgQQAIg==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDCfMNsN6miEE3rVyUPwElfiJSWaR5huPCzUenZOfJT04GAcQdWvEju3
UM2lmBLIXpGgBwYFK4EEACKhZANiAARCjRMqJ85rzMC998X5z761nJ+xL3bkmGVq
WvrJ51t5OxV0v25NsOgR82CANXUgvhVYs7vNFN+jxtb2aj6Xg+/2G/BNxkaFspIV
CzgWkxiz7XE4lgUwX44FCXZM3+JeUbI=
-----END EC PRIVATE KEY-----`
	caCRT = `-----BEGIN CERTIFICATE-----
MIIE5jCCAs6gAwIBAgIBATANBgkqhkiG9w0BAQsFADATMREwDwYDVQQDEwhDZXJ0
QXV0aDAeFw0yMTAxMDIyMTIwNTVaFw0yMjA3MDIyMTMwNTJaMBMxETAPBgNVBAMT
CENlcnRBdXRoMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA4Tiho5xW
AC15JRkMwfp3/TJwI2As7MY5dele5cmdr5bHAE+sRKqC+Ti88OJWCV5saoyax/1S
CjxJlQMZMl169P1QYJskKjdG2sdv6RLWLMgwSNRRjxp/Bw9dHdiEb9MjLgu28Jro
9peQkHcRHeMf5hM9WvlIJGrdzbC4hUehmqggcqgARainBkYjf0SwuWxHeu4nMqkp
Ak5tcSTLCjHfEFHZ9Te0TIPG5YkWocQKyeLgu4lvuU+DD2W2lym+YVUtRMGs1Env
k7p+N0DcGU26qfzZ2sF5ZXkqm7dBsGQB9pIxwc2Q8T1dCIyP9OQCKVILdc5aVFf1
cryQFHYzYNNZXFlIBims5VV5Mgfp8ESHQSue+v6n6ykecLEyKt1F1Y/MWY/nWUSI
8zdq83jdBAZVjo9MSthxVn57/06s/hQca65IpcTZV2gX0a+eRlAVqaRbAhL3LaZe
bYsW3WHKoUOftwemuep3nL51TzlXZVL7Oz/ClGaEOsnGG9KFO6jh+W768qC0zLQI
CdE7v2Zex98sZteHCg9fGJHIaYoF0aJG5P3WI5oZf2fy7UIYN9ADLFZiorCXAZEh
CSU6mDoRViZ4RGR9GZxbDZ9KYn7O8M/KCR72bkQg73TlMsk1zSXEw0MKLUjtsw6c
rZ0Jt8t3sRatHO3JrYHALMt9vZfyNCZp0IsCAwEAAaNFMEMwDgYDVR0PAQH/BAQD
AgEGMBIGA1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0OBBYEFO1yCNAGr/zQTJIi8lw3
w5OiuBvMMA0GCSqGSIb3DQEBCwUAA4ICAQA6gCNuM7r8mnx674dm31GxBjQy5ZwB
7CxDzYEvL/oiZ3Tv3HlPfN2LAAsJUfGnghh9DOytenL2CTZWjl/emP5eijzmlP+9
zva5I6CIMCf/eDDVsRdO244t0o4uG7+At0IgSDM3bpVaVb4RHZNjEziYChsEYY8d
HK6iwuRSvFniV6yhR/Vj1Ymi9yZ5xclqseLXiQnUB0PkfIk23+7s42cXB16653fH
O/FsPyKBLiKJArizLYQc12aP3QOrYoYD9+fAzIIzew7A5C0aanZCGzkuFpO6TRlD
Tb7ry9Gf0DfPpCgxraH8tOcmnqp/ka3hjqo/SRnnTk0IFrmmLdarJvjD46rKwBo4
MjyAIR1mQ5j8GTlSFBmSgETOQ/EYvO3FPLmra1Fh7L+DvaVzTpqI9fG3TuyyY+Ri
Fby4ycTOGSZOe5Fh8lqkX5Y47mCUJ3zHzOA1vUJy2eTlMRGpu47Eb1++Vm6EzPUP
2EF5aD+zwcssh+atZvQbwxpgVqVcyLt91RSkKkmZQslh0rnlTb68yxvUnD3zw7So
o6TAf9UvwVMEvdLT9NnFd6hwi2jcNte/h538GJwXeBb8EkfpqLKpTKyicnOdkamZ
7E9zY8SHNRYMwB9coQ/W8NvufbCgkvOoLyMXk5edbXofXl3PhNGOlraWbghBnzf5
r3rwjFsQOoZotA==
-----END CERTIFICATE-----`
	caKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`
	caCRL = `-----BEGIN X509 CRL-----
MIICpzCBkAIBATANBgkqhkiG9w0BAQsFADATMREwDwYDVQQDEwhDZXJ0QXV0aBcN
MjEwMTAyMjEzNDA1WhcNMjMwMTAyMjEzNDA1WjAkMCICEQC+l04DbHWMyC3fG09k
VXf+Fw0yMTAxMDIyMTM0MDVaoCMwITAfBgNVHSMEGDAWgBTtcgjQBq/80EySIvJc
N8OTorgbzDANBgkqhkiG9w0BAQsFAAOCAgEAEJ7z+uNc8sqtxlOhSdTGDzX/xput
E857kFQkSlMnU2whQ8c+XpYrBLA5vIZJNSSwohTpM4+zVBX/bJpmu3wqqaArRO9/
YcW5mQk9Anvb4WjQW1cHmtNapMTzoC9AiYt/OWPfy+P6JCgCr4Hy6LgQyIRL6bM9
VYTalolOm1qa4Y5cIeT7iHq/91mfaqo8/6MYRjLl8DOTROpmw8OS9bCXkzGKdCat
AbAzwkQUSauyoCQ10rpX+Y64w9ng3g4Dr20aCqPf5osaqplEJ2HTK8ljDTidlslv
9anQj8ax3Su89vI8+hK+YbfVQwrThabgdSjQsn+veyx8GlP8WwHLAQ379KjZjWg+
OlOSwBeU1vTdP0QcB8X5C2gVujAyuQekbaV86xzIBOj7vZdfHZ6ee30TZ2FKiMyg
7/N2OqW0w77ChsjB4MSHJCfuTgIeg62GzuZXLM+Q2Z9LBdtm4Byg+sm/P52adOEg
gVb2Zf4KSvsAmA0PIBlu449/QXUFcMxzLFy7mwTeZj2B4Ln0Hm0szV9f9R8MwMtB
SyLYxVH+mgqaR6Jkk22Q/yYyLPaELfafX5gp/AIXG8n0zxfVaTvK3auSgb1Q6ZLS
5QH9dSIsmZHlPq7GoSXmKpMdjUL8eaky/IMteioyXgsBiATzl5L2dsw6MTX3MDF0
QbDK+MzhmbKfDxs=
-----END X509 CRL-----`
	client1Crt = `-----BEGIN CERTIFICATE-----
MIIEITCCAgmgAwIBAgIRAIppZHoj1hM80D7WzTEKLuAwDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEAxMIQ2VydEF1dGgwHhcNMjEwMTAyMjEyMzEwWhcNMjIwNzAyMjEz
MDUxWjASMRAwDgYDVQQDEwdjbGllbnQxMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAoKbYY9MdF2kF/nhBESIiZTdVYtA8XL9xrIZyDj9EnCiTxHiVbJtH
XVwszqSl5TRrotPmnmAQcX3r8OCk+z+RQZ0QQj257P3kG6q4rNnOcWCS5xEd20jP
yhQ3m+hMGfZsotNTQze1ochuQgLUN6IPyPxZkH22ia3jX4iu1eo/QxeLYHj1UHw4
3Cii9yE+j5kPUC21xmnrGKdUrB55NYLXHx6yTIqYR5znSOVB8oJi18/hwdZmH859
DHhm0Hx1HrS+jbjI3+CMorZJ3WUyNf+CkiVLD3xYutPbxzEpwiqkG/XYzLH0habT
cDcILo18n+o3jvem2KWBrDhyairjIDscwQIDAQABo3EwbzAOBgNVHQ8BAf8EBAMC
A7gwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBSJ5GIv
zIrE4ZSQt2+CGblKTDswizAfBgNVHSMEGDAWgBTtcgjQBq/80EySIvJcN8OTorgb
zDANBgkqhkiG9w0BAQsFAAOCAgEALh4f5GhvNYNou0Ab04iQBbLEdOu2RlbK1B5n
K9P/umYenBHMY/z6HT3+6tpcHsDuqE8UVdq3f3Gh4S2Gu9m8PRitT+cJ3gdo9Plm
3rD4ufn/s6rGg3ppydXcedm17492tbccUDWOBZw3IO/ASVq13WPgT0/Kev7cPq0k
sSdSNhVeXqx8Myc2/d+8GYyzbul2Kpfa7h9i24sK49E9ftnSmsIvngONo08eT1T0
3wAOyK2981LIsHaAWcneShKFLDB6LeXIT9oitOYhiykhFlBZ4M1GNlSNfhQ8IIQP
xbqMNXCLkW4/BtLhGEEcg0QVso6Kudl9rzgTfQknrdF7pHp6rS46wYUjoSyIY6dl
oLmnoAVJX36J3QPWelePI9e07X2wrTfiZWewwgw3KNRWjd6/zfPLe7GoqXnK1S2z
PT8qMfCaTwKTtUkzXuTFvQ8bAo2My/mS8FOcpkt2oQWeOsADHAUX7fz5BCoa2DL3
k/7Mh4gVT+JYZEoTwCFuYHgMWFWe98naqHi9lB4yR981p1QgXgxO7qBeipagKY1F
LlH1iwXUqZ3MZnkNA+4e1Fglsw3sa/rC+L98HnznJ/YbTfQbCP6aQ1qcOymrjMud
7MrFwqZjtd/SK4Qx1VpK6jGEAtPgWBTUS3p9ayg6lqjMBjsmySWfvRsDQbq6P5Ct
O/e3EH8=
-----END CERTIFICATE-----`
	client1Key = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAoKbYY9MdF2kF/nhBESIiZTdVYtA8XL9xrIZyDj9EnCiTxHiV
bJtHXVwszqSl5TRrotPmnmAQcX3r8OCk+z+RQZ0QQj257P3kG6q4rNnOcWCS5xEd
20jPyhQ3m+hMGfZsotNTQze1ochuQgLUN6IPyPxZkH22ia3jX4iu1eo/QxeLYHj1
UHw43Cii9yE+j5kPUC21xmnrGKdUrB55NYLXHx6yTIqYR5znSOVB8oJi18/hwdZm
H859DHhm0Hx1HrS+jbjI3+CMorZJ3WUyNf+CkiVLD3xYutPbxzEpwiqkG/XYzLH0
habTcDcILo18n+o3jvem2KWBrDhyairjIDscwQIDAQABAoIBAEBSjVFqtbsp0byR
aXvyrtLX1Ng7h++at2jca85Ihq//jyqbHTje8zPuNAKI6eNbmb0YGr5OuEa4pD9N
ssDmMsKSoG/lRwwcm7h4InkSvBWpFShvMgUaohfHAHzsBYxfnh+TfULsi0y7c2n6
t/2OZcOTRkkUDIITnXYiw93ibHHv2Mv2bBDu35kGrcK+c2dN5IL5ZjTjMRpbJTe2
44RBJbdTxHBVSgoGBnugF+s2aEma6Ehsj70oyfoVpM6Aed5kGge0A5zA1JO7WCn9
Ay/DzlULRXHjJIoRWd2NKvx5n3FNppUc9vJh2plRHalRooZ2+MjSf8HmXlvG2Hpb
ScvmWgECgYEA1G+A/2KnxWsr/7uWIJ7ClcGCiNLdk17Pv3DZ3G4qUsU2ITftfIbb
tU0Q/b19na1IY8Pjy9ptP7t74/hF5kky97cf1FA8F+nMj/k4+wO8QDI8OJfzVzh9
PwielA5vbE+xmvis5Hdp8/od1Yrc/rPSy2TKtPFhvsqXjqoUmOAjDP8CgYEAwZjH
9dt1sc2lx/rMxihlWEzQ3JPswKW9/LJAmbRBoSWF9FGNjbX7uhWtXRKJkzb8ZAwa
88azluNo2oftbDD/+jw8b2cDgaJHlLAkSD4O1D1RthW7/LKD15qZ/oFsRb13NV85
ZNKtwslXGbfVNyGKUVFm7fVA8vBAOUey+LKDFj8CgYEAg8WWstOzVdYguMTXXuyb
ruEV42FJaDyLiSirOvxq7GTAKuLSQUg1yMRBIeQEo2X1XU0JZE3dLodRVhuO4EXP
g7Dn4X7Th9HSvgvNuIacowWGLWSz4Qp9RjhGhXhezUSx2nseY6le46PmFavJYYSR
4PBofMyt4PcyA6Cknh+KHmkCgYEAnTriG7ETE0a7v4DXUpB4TpCEiMCy5Xs2o8Z5
ZNva+W+qLVUWq+MDAIyechqeFSvxK6gRM69LJ96lx+XhU58wJiFJzAhT9rK/g+jS
bsHH9WOfu0xHkuHA5hgvvV2Le9B2wqgFyva4HJy82qxMxCu/VG/SMqyfBS9OWbb7
ibQhdq0CgYAl53LUWZsFSZIth1vux2LVOsI8C3X1oiXDGpnrdlQ+K7z57hq5EsRq
GC+INxwXbvKNqp5h0z2MvmKYPDlGVTgw8f8JjM7TkN17ERLcydhdRrMONUryZpo8
1xTob+8blyJgfxZUIAKbMbMbIiU0WAF0rfD/eJJwS4htOW/Hfv4TGA==
-----END RSA PRIVATE KEY-----`
	// client 2 crt is revoked
	client2Crt = `-----BEGIN CERTIFICATE-----
MIIEITCCAgmgAwIBAgIRAL6XTgNsdYzILd8bT2RVd/4wDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEAxMIQ2VydEF1dGgwHhcNMjEwMTAyMjEyMzIwWhcNMjIwNzAyMjEz
MDUxWjASMRAwDgYDVQQDEwdjbGllbnQyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEA6xjW5KQR3/OFQtV5M75WINqQ4AzXSu6DhSz/yumaaQZP/UxY+6hi
jcrFzGo9MMie/Sza8DhkXOFAl2BelUubrOeB2cl+/Gr8OCyRi2Gv6j3zCsuN/4jQ
tNaoez/IbkDvI3l/ZpzBtnuNY2RiemGgHuORXHRVf3qVlsw+npBIRW5rM2HkO/xG
oZjeBErWVu390Lyn+Gvk2TqQDnkutWnxUC60/zPlHhXZ4BwaFAekbSnjsSDB1YFM
s8HwW4oBryoxdj3/+/qLrBHt75IdLw3T7/V1UDJQM3EvSQOr12w4egpldhtsC871
nnBQZeY6qA5feffIwwg/6lJm70o6S6OX6wIDAQABo3EwbzAOBgNVHQ8BAf8EBAMC
A7gwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBTB84v5
t9HqhLhMODbn6oYkEQt3KzAfBgNVHSMEGDAWgBTtcgjQBq/80EySIvJcN8OTorgb
zDANBgkqhkiG9w0BAQsFAAOCAgEALGtBCve5k8tToL3oLuXp/oSik6ovIB/zq4I/
4zNMYPU31+ZWz6aahysgx1JL1yqTa3Qm8o2tu52MbnV10dM7CIw7c/cYa+c+OPcG
5LF97kp13X+r2axy+CmwM86b4ILaDGs2Qyai6VB6k7oFUve+av5o7aUrNFpqGCJz
HWdtHZSVA3JMATzy0TfWanwkzreqfdw7qH0yZ9bDURlBKAVWrqnCstva9jRuv+AI
eqxr/4Ro986TFjJdoAP3Vr16CPg7/B6GA/KmsBWJrpeJdPWq4i2gpLKvYZoy89qD
mUZf34RbzcCtV4NvV1DadGnt4us0nvLrvS5rL2+2uWD09kZYq9RbLkvgzF/cY0fz
i7I1bi5XQ+alWe0uAk5ZZL/D+GTRYUX1AWwCqwJxmHrMxcskMyO9pXvLyuSWRDLo
YNBrbX9nLcfJzVCp+X+9sntTHjs4l6Cw+fLepJIgtgqdCHtbhTiv68vSM6cgb4br
6n2xrXRKuioiWFOrTSRr+oalZh8dGJ/xvwY8IbWknZAvml9mf1VvfE7Ma5P777QM
fsbYVTq0Y3R/5hIWsC3HA5z6MIM8L1oRe/YyhP3CTmrCHkVKyDOosGXpGz+JVcyo
cfYkY5A3yFKB2HaCwZSfwFmRhxkrYWGEbHv3Cd9YkZs1J3hNhGFZyVMC9Uh0S85a
6zdDidU=
-----END CERTIFICATE-----`
	client2Key = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA6xjW5KQR3/OFQtV5M75WINqQ4AzXSu6DhSz/yumaaQZP/UxY
+6hijcrFzGo9MMie/Sza8DhkXOFAl2BelUubrOeB2cl+/Gr8OCyRi2Gv6j3zCsuN
/4jQtNaoez/IbkDvI3l/ZpzBtnuNY2RiemGgHuORXHRVf3qVlsw+npBIRW5rM2Hk
O/xGoZjeBErWVu390Lyn+Gvk2TqQDnkutWnxUC60/zPlHhXZ4BwaFAekbSnjsSDB
1YFMs8HwW4oBryoxdj3/+/qLrBHt75IdLw3T7/V1UDJQM3EvSQOr12w4egpldhts
C871nnBQZeY6qA5feffIwwg/6lJm70o6S6OX6wIDAQABAoIBAFatstVb1KdQXsq0
cFpui8zTKOUiduJOrDkWzTygAmlEhYtrccdfXu7OWz0x0lvBLDVGK3a0I/TGrAzj
4BuFY+FM/egxTVt9in6fmA3et4BS1OAfCryzUdfK6RV//8L+t+zJZ/qKQzWnugpy
QYjDo8ifuMFwtvEoXizaIyBNLAhEp9hnrv+Tyi2O2gahPvCHsD48zkyZRCHYRstD
NH5cIrwz9/RJgPO1KI+QsJE7Nh7stR0sbr+5TPU4fnsL2mNhMUF2TJrwIPrc1yp+
YIUjdnh3SO88j4TQT3CIrWi8i4pOy6N0dcVn3gpCRGaqAKyS2ZYUj+yVtLO4KwxZ
SZ1lNvECgYEA78BrF7f4ETfWSLcBQ3qxfLs7ibB6IYo2x25685FhZjD+zLXM1AKb
FJHEXUm3mUYrFJK6AFEyOQnyGKBOLs3S6oTAswMPbTkkZeD1Y9O6uv0AHASLZnK6
pC6ub0eSRF5LUyTQ55Jj8D7QsjXJueO8v+G5ihWhNSN9tB2UA+8NBmkCgYEA+weq
cvoeMIEMBQHnNNLy35bwfqrceGyPIRBcUIvzQfY1vk7KW6DYOUzC7u+WUzy/hA52
DjXVVhua2eMQ9qqtOav7djcMc2W9RbLowxvno7K5qiCss013MeWk64TCWy+WMp5A
AVAtOliC3hMkIKqvR2poqn+IBTh1449agUJQqTMCgYEAu06IHGq1GraV6g9XpGF5
wqoAlMzUTdnOfDabRilBf/YtSr+J++ThRcuwLvXFw7CnPZZ4TIEjDJ7xjj3HdxeE
fYYjineMmNd40UNUU556F1ZLvJfsVKizmkuCKhwvcMx+asGrmA+tlmds4p3VMS50
KzDtpKzLWlmU/p/RINWlRmkCgYBy0pHTn7aZZx2xWKqCDg+L2EXPGqZX6wgZDpu7
OBifzlfM4ctL2CmvI/5yPmLbVgkgBWFYpKUdiujsyyEiQvWTUKhn7UwjqKDHtcsk
G6p7xS+JswJrzX4885bZJ9Oi1AR2yM3sC9l0O7I4lDbNPmWIXBLeEhGMmcPKv/Kc
91Ff4wKBgQCF3ur+Vt0PSU0ucrPVHjCe7tqazm0LJaWbPXL1Aw0pzdM2EcNcW/MA
w0kqpr7MgJ94qhXCBcVcfPuFN9fBOadM3UBj1B45Cz3pptoK+ScI8XKno6jvVK/p
xr5cb9VBRBtB9aOKVfuRhpatAfS2Pzm2Htae9lFn7slGPUmu2hkjDw==
-----END RSA PRIVATE KEY-----`
)

func TestShouldBind(t *testing.T) {
	c := Conf{
		Bindings: []Binding{
			{
				Port: 10000,
			},
		},
	}
	require.True(t, c.ShouldBind())

	c.Bindings[0].Port = 0
	require.False(t, c.ShouldBind())

	if runtime.GOOS != osWindows {
		c.Bindings[0].Address = "/absolute/path"
		require.True(t, c.ShouldBind())
	}
}

func TestGetRespStatus(t *testing.T) {
	var err error
	err = &dataprovider.MethodDisabledError{}
	respStatus := getRespStatus(err)
	assert.Equal(t, http.StatusForbidden, respStatus)
	err = fmt.Errorf("generic error")
	respStatus = getRespStatus(err)
	assert.Equal(t, http.StatusInternalServerError, respStatus)
}

func TestGCSWebInvalidFormFile(t *testing.T) {
	form := make(url.Values)
	form.Set("username", "test_username")
	form.Set("fs_provider", "2")
	req, _ := http.NewRequest(http.MethodPost, webUserPath, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err := req.ParseForm()
	assert.NoError(t, err)
	_, err = getFsConfigFromUserPostFields(req)
	assert.EqualError(t, err, http.ErrNotMultipart.Error())
}

func TestInvalidToken(t *testing.T) {
	admin := dataprovider.Admin{
		Username: "admin",
	}
	errFake := errors.New("fake error")
	asJSON, err := json.Marshal(admin)
	assert.NoError(t, err)
	req, _ := http.NewRequest(http.MethodPut, path.Join(adminPath, admin.Username), bytes.NewBuffer(asJSON))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("username", admin.Username)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), jwtauth.ErrorCtxKey, errFake))
	rr := httptest.NewRecorder()
	updateAdmin(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	rr = httptest.NewRecorder()
	deleteAdmin(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	adminPwd := adminPwd{
		CurrentPassword: "old",
		NewPassword:     "new",
	}
	asJSON, err = json.Marshal(adminPwd)
	assert.NoError(t, err)
	req, _ = http.NewRequest(http.MethodPut, "", bytes.NewBuffer(asJSON))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), jwtauth.ErrorCtxKey, errFake))
	rr = httptest.NewRecorder()
	changeAdminPassword(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	adm := getAdminFromToken(req)
	assert.Empty(t, adm.Username)
}

func TestUpdateWebAdminInvalidClaims(t *testing.T) {
	server := httpdServer{}
	server.initializeRouter()

	rr := httptest.NewRecorder()
	admin := dataprovider.Admin{
		Username: "",
		Password: "password",
	}
	c := jwtTokenClaims{
		Username:    admin.Username,
		Permissions: admin.Permissions,
		Signature:   admin.GetSignature(),
	}
	token, err := c.createTokenResponse(server.tokenAuth, tokenAudienceWeb)
	assert.NoError(t, err)

	form := make(url.Values)
	form.Set("status", "1")
	req, _ := http.NewRequest(http.MethodPost, path.Join(webAdminPath, "admin"), bytes.NewBuffer([]byte(form.Encode())))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("username", "admin")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", fmt.Sprintf("jwt=%v", token["access_token"]))
	handleWebUpdateAdminPost(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid token claims")
}

func TestCreateTokenError(t *testing.T) {
	server := httpdServer{
		tokenAuth: jwtauth.New("PS256", utils.GenerateRandomBytes(32), nil),
	}
	rr := httptest.NewRecorder()
	admin := dataprovider.Admin{
		Username: "admin",
		Password: "password",
	}
	req, _ := http.NewRequest(http.MethodGet, tokenPath, nil)

	server.checkAddrAndSendToken(rr, req, admin)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	rr = httptest.NewRecorder()
	form := make(url.Values)
	form.Set("username", admin.Username)
	form.Set("password", admin.Password)
	req, _ = http.NewRequest(http.MethodPost, webLoginPath, bytes.NewBuffer([]byte(form.Encode())))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.handleWebLoginPost(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	// req with no content type
	req, _ = http.NewRequest(http.MethodPost, webLoginPath, nil)
	rr = httptest.NewRecorder()
	server.handleWebLoginPost(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	// req with no POST body
	req, _ = http.NewRequest(http.MethodGet, webLoginPath+"?a=a%C3%AO%GG", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	server.handleWebLoginPost(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	req, _ = http.NewRequest(http.MethodGet, webLoginPath+"?a=a%C3%A1%G2", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handleWebAdminChangePwdPost(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	req, _ = http.NewRequest(http.MethodGet, webLoginPath+"?a=a%C3%A2%G3", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, err := getAdminFromPostFields(req)
	assert.Error(t, err)
}

func TestJWTTokenValidation(t *testing.T) {
	tokenAuth := jwtauth.New("HS256", utils.GenerateRandomBytes(32), nil)
	claims := make(map[string]interface{})
	claims["username"] = "admin"
	claims[jwt.ExpirationKey] = time.Now().UTC().Add(-1 * time.Hour)
	token, _, err := tokenAuth.Encode(claims)
	assert.NoError(t, err)

	r := GetHTTPRouter()
	fn := jwtAuthenticator(r)
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, userPath, nil)
	ctx := jwtauth.NewContext(req.Context(), token, nil)
	fn.ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	fn = jwtAuthenticatorWeb(r)
	rr = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, webUserPath, nil)
	ctx = jwtauth.NewContext(req.Context(), token, nil)

	fn.ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusFound, rr.Code)

	errTest := errors.New("test error")
	permFn := checkPerm(dataprovider.PermAdminAny)
	fn = permFn(r)
	rr = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, userPath, nil)
	ctx = jwtauth.NewContext(req.Context(), token, errTest)
	fn.ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	permFn = checkPerm(dataprovider.PermAdminAny)
	fn = permFn(r)
	rr = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, webUserPath, nil)
	req.RequestURI = webUserPath
	ctx = jwtauth.NewContext(req.Context(), token, errTest)
	fn.ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminAllowListConnAddr(t *testing.T) {
	server := httpdServer{}
	admin := dataprovider.Admin{
		Filters: dataprovider.AdminFilters{
			AllowList: []string{"192.168.1.0/24"},
		},
	}
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, tokenPath, nil)
	ctx := context.WithValue(req.Context(), connAddrKey, "127.0.0.1:4567")
	req.RemoteAddr = "192.168.1.16:1234"
	server.checkAddrAndSendToken(rr, req.WithContext(ctx), admin)
	assert.Equal(t, http.StatusForbidden, rr.Code, rr.Body.String())
}

func TestUpdateContextFromCookie(t *testing.T) {
	server := httpdServer{
		tokenAuth: jwtauth.New("HS256", utils.GenerateRandomBytes(32), nil),
	}
	req, _ := http.NewRequest(http.MethodGet, tokenPath, nil)
	claims := make(map[string]interface{})
	claims["a"] = "b"
	token, _, err := server.tokenAuth.Encode(claims)
	assert.NoError(t, err)

	ctx := jwtauth.NewContext(req.Context(), token, nil)
	server.updateContextFromCookie(req.WithContext(ctx))
}

func TestCookieExpiration(t *testing.T) {
	server := httpdServer{
		tokenAuth: jwtauth.New("HS256", utils.GenerateRandomBytes(32), nil),
	}
	err := errors.New("test error")
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, tokenPath, nil)
	ctx := jwtauth.NewContext(req.Context(), nil, err)
	server.checkCookieExpiration(rr, req.WithContext(ctx))
	cookie := rr.Header().Get("Set-Cookie")
	assert.Empty(t, cookie)

	req, _ = http.NewRequest(http.MethodGet, tokenPath, nil)
	claims := make(map[string]interface{})
	claims["a"] = "b"
	token, _, err := server.tokenAuth.Encode(claims)
	assert.NoError(t, err)
	ctx = jwtauth.NewContext(req.Context(), token, nil)
	server.checkCookieExpiration(rr, req.WithContext(ctx))
	cookie = rr.Header().Get("Set-Cookie")
	assert.Empty(t, cookie)

	admin := dataprovider.Admin{
		Username:    "newtestadmin",
		Password:    "password",
		Permissions: []string{dataprovider.PermAdminAny},
	}
	claims = make(map[string]interface{})
	claims[claimUsernameKey] = admin.Username
	claims[claimPermissionsKey] = admin.Permissions
	claims[jwt.SubjectKey] = admin.GetSignature()
	claims[jwt.ExpirationKey] = time.Now().Add(1 * time.Minute)
	token, _, err = server.tokenAuth.Encode(claims)
	assert.NoError(t, err)
	req, _ = http.NewRequest(http.MethodGet, tokenPath, nil)
	ctx = jwtauth.NewContext(req.Context(), token, nil)
	server.checkCookieExpiration(rr, req.WithContext(ctx))
	cookie = rr.Header().Get("Set-Cookie")
	assert.Empty(t, cookie)

	admin.Status = 0
	err = dataprovider.AddAdmin(&admin)
	assert.NoError(t, err)
	req, _ = http.NewRequest(http.MethodGet, tokenPath, nil)
	ctx = jwtauth.NewContext(req.Context(), token, nil)
	server.checkCookieExpiration(rr, req.WithContext(ctx))
	cookie = rr.Header().Get("Set-Cookie")
	assert.Empty(t, cookie)

	admin.Status = 1
	admin.Filters.AllowList = []string{"172.16.1.0/24"}
	err = dataprovider.UpdateAdmin(&admin)
	assert.NoError(t, err)
	req, _ = http.NewRequest(http.MethodGet, tokenPath, nil)
	ctx = jwtauth.NewContext(req.Context(), token, nil)
	server.checkCookieExpiration(rr, req.WithContext(ctx))
	cookie = rr.Header().Get("Set-Cookie")
	assert.Empty(t, cookie)

	admin, err = dataprovider.AdminExists(admin.Username)
	assert.NoError(t, err)
	claims = make(map[string]interface{})
	claims[claimUsernameKey] = admin.Username
	claims[claimPermissionsKey] = admin.Permissions
	claims[jwt.SubjectKey] = admin.GetSignature()
	claims[jwt.ExpirationKey] = time.Now().Add(1 * time.Minute)
	token, _, err = server.tokenAuth.Encode(claims)
	assert.NoError(t, err)
	req, _ = http.NewRequest(http.MethodGet, tokenPath, nil)
	req.RemoteAddr = "192.168.8.1:1234"
	ctx = jwtauth.NewContext(req.Context(), token, nil)
	server.checkCookieExpiration(rr, req.WithContext(ctx))
	cookie = rr.Header().Get("Set-Cookie")
	assert.Empty(t, cookie)

	req, _ = http.NewRequest(http.MethodGet, tokenPath, nil)
	req.RemoteAddr = "172.16.1.2:1234"
	ctx = jwtauth.NewContext(req.Context(), token, nil)
	ctx = context.WithValue(ctx, connAddrKey, "10.9.9.9")
	server.checkCookieExpiration(rr, req.WithContext(ctx))
	cookie = rr.Header().Get("Set-Cookie")
	assert.Empty(t, cookie)

	req, _ = http.NewRequest(http.MethodGet, tokenPath, nil)
	req.RemoteAddr = "172.16.1.12:4567"
	ctx = jwtauth.NewContext(req.Context(), token, nil)
	server.checkCookieExpiration(rr, req.WithContext(ctx))
	cookie = rr.Header().Get("Set-Cookie")
	assert.True(t, strings.HasPrefix(cookie, "jwt="))

	err = dataprovider.DeleteAdmin(admin.Username)
	assert.NoError(t, err)
}

func TestGetURLParam(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, adminPwdPath, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("val", "testuser%C3%A0")
	rctx.URLParams.Add("inval", "testuser%C3%AO%GG")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	escaped := getURLParam(req, "val")
	assert.Equal(t, "testuserà", escaped)
	escaped = getURLParam(req, "inval")
	assert.Equal(t, "testuser%C3%AO%GG", escaped)
}

func TestChangePwdValidationErrors(t *testing.T) {
	err := doChangeAdminPassword(nil, "", "", "")
	require.Error(t, err)
	err = doChangeAdminPassword(nil, "a", "b", "c")
	require.Error(t, err)
	err = doChangeAdminPassword(nil, "a", "a", "a")
	require.Error(t, err)

	req, _ := http.NewRequest(http.MethodPut, adminPwdPath, nil)
	err = doChangeAdminPassword(req, "currentpwd", "newpwd", "newpwd")
	assert.Error(t, err)
}

func TestRenderUnexistingFolder(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, folderPath, nil)
	renderFolder(rr, req, "path not mapped", http.StatusOK)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCloseConnectionHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodDelete, activeConnectionsPath+"/connectionID", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("connectionID", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()
	handleCloseConnection(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRenderInvalidTemplate(t *testing.T) {
	tmpl, err := template.New("test").Parse("{{.Count}}")
	if assert.NoError(t, err) {
		templates["no_match"] = tmpl
		rw := httptest.NewRecorder()
		renderTemplate(rw, "no_match", map[string]string{})
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	}
}

func TestQuotaScanInvalidFs(t *testing.T) {
	user := dataprovider.User{
		Username: "test",
		HomeDir:  os.TempDir(),
		FsConfig: dataprovider.Filesystem{
			Provider: dataprovider.S3FilesystemProvider,
		},
	}
	common.QuotaScans.AddUserQuotaScan(user.Username)
	err := doQuotaScan(user)
	assert.Error(t, err)
}

func TestVerifyTLSConnection(t *testing.T) {
	oldCertMgr := certMgr

	caCrlPath := filepath.Join(os.TempDir(), "testcrl.crt")
	certPath := filepath.Join(os.TempDir(), "testh.crt")
	keyPath := filepath.Join(os.TempDir(), "testh.key")
	err := ioutil.WriteFile(caCrlPath, []byte(caCRL), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(certPath, []byte(httpdCert), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(keyPath, []byte(httpdKey), os.ModePerm)
	assert.NoError(t, err)

	certMgr, err = common.NewCertManager(certPath, keyPath, "", "webdav_test")
	assert.NoError(t, err)

	certMgr.SetCARevocationLists([]string{caCrlPath})
	err = certMgr.LoadCRLs()
	assert.NoError(t, err)

	crt, err := tls.X509KeyPair([]byte(client1Crt), []byte(client1Key))
	assert.NoError(t, err)
	x509crt, err := x509.ParseCertificate(crt.Certificate[0])
	assert.NoError(t, err)

	server := httpdServer{}
	state := tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{x509crt},
	}

	err = server.verifyTLSConnection(state)
	assert.Error(t, err) // no verified certification chain

	crt, err = tls.X509KeyPair([]byte(caCRT), []byte(caKey))
	assert.NoError(t, err)

	x509CAcrt, err := x509.ParseCertificate(crt.Certificate[0])
	assert.NoError(t, err)

	state.VerifiedChains = append(state.VerifiedChains, []*x509.Certificate{x509crt, x509CAcrt})
	err = server.verifyTLSConnection(state)
	assert.NoError(t, err)

	crt, err = tls.X509KeyPair([]byte(client2Crt), []byte(client2Key))
	assert.NoError(t, err)
	x509crtRevoked, err := x509.ParseCertificate(crt.Certificate[0])
	assert.NoError(t, err)

	state.VerifiedChains = append(state.VerifiedChains, []*x509.Certificate{x509crtRevoked, x509CAcrt})
	state.PeerCertificates = []*x509.Certificate{x509crtRevoked}
	err = server.verifyTLSConnection(state)
	assert.EqualError(t, err, common.ErrCrtRevoked.Error())

	err = os.Remove(caCrlPath)
	assert.NoError(t, err)
	err = os.Remove(certPath)
	assert.NoError(t, err)
	err = os.Remove(keyPath)
	assert.NoError(t, err)

	certMgr = oldCertMgr
}

func TestGetUserFromTemplate(t *testing.T) {
	user := dataprovider.User{
		Status: 1,
	}
	user.VirtualFolders = append(user.VirtualFolders, vfs.VirtualFolder{
		BaseVirtualFolder: vfs.BaseVirtualFolder{
			Name: "Folder%username%",
		},
	})

	username := "userTemplate"
	password := "pwdTemplate"
	templateFields := userTemplateFields{
		Username: username,
		Password: password,
	}

	userTemplate := getUserFromTemplate(user, templateFields)
	require.Len(t, userTemplate.VirtualFolders, 1)
	require.Equal(t, "Folder"+username, userTemplate.VirtualFolders[0].Name)

	user.FsConfig.Provider = dataprovider.CryptedFilesystemProvider
	user.FsConfig.CryptConfig.Passphrase = kms.NewPlainSecret("%password%")
	userTemplate = getUserFromTemplate(user, templateFields)
	require.Equal(t, password, userTemplate.FsConfig.CryptConfig.Passphrase.GetPayload())

	user.FsConfig.Provider = dataprovider.GCSFilesystemProvider
	user.FsConfig.GCSConfig.KeyPrefix = "%username%%password%"
	userTemplate = getUserFromTemplate(user, templateFields)
	require.Equal(t, username+password, userTemplate.FsConfig.GCSConfig.KeyPrefix)

	user.FsConfig.Provider = dataprovider.AzureBlobFilesystemProvider
	user.FsConfig.AzBlobConfig.KeyPrefix = "a%username%"
	user.FsConfig.AzBlobConfig.AccountKey = kms.NewPlainSecret("pwd%password%%username%")
	userTemplate = getUserFromTemplate(user, templateFields)
	require.Equal(t, "a"+username, userTemplate.FsConfig.AzBlobConfig.KeyPrefix)
	require.Equal(t, "pwd"+password+username, userTemplate.FsConfig.AzBlobConfig.AccountKey.GetPayload())

	user.FsConfig.Provider = dataprovider.SFTPFilesystemProvider
	user.FsConfig.SFTPConfig.Prefix = "%username%"
	user.FsConfig.SFTPConfig.Username = "sftp_%username%"
	user.FsConfig.SFTPConfig.Password = kms.NewPlainSecret("sftp%password%")
	userTemplate = getUserFromTemplate(user, templateFields)
	require.Equal(t, username, userTemplate.FsConfig.SFTPConfig.Prefix)
	require.Equal(t, "sftp_"+username, userTemplate.FsConfig.SFTPConfig.Username)
	require.Equal(t, "sftp"+password, userTemplate.FsConfig.SFTPConfig.Password.GetPayload())
}

func TestJWTTokenCleanup(t *testing.T) {
	server := httpdServer{
		tokenAuth: jwtauth.New("HS256", utils.GenerateRandomBytes(32), nil),
	}
	admin := dataprovider.Admin{
		Username:    "newtestadmin",
		Password:    "password",
		Permissions: []string{dataprovider.PermAdminAny},
	}
	claims := make(map[string]interface{})
	claims[claimUsernameKey] = admin.Username
	claims[claimPermissionsKey] = admin.Permissions
	claims[jwt.SubjectKey] = admin.GetSignature()
	claims[jwt.ExpirationKey] = time.Now().Add(1 * time.Minute)
	_, token, err := server.tokenAuth.Encode(claims)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, versionPath, nil)
	assert.True(t, isTokenInvalidated(req))

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

	invalidatedJWTTokens.Store(token, time.Now().UTC().Add(-tokenDuration))
	require.True(t, isTokenInvalidated(req))
	startJWTTokensCleanupTicker(100 * time.Millisecond)
	assert.Eventually(t, func() bool { return !isTokenInvalidated(req) }, 1*time.Second, 200*time.Millisecond)
	stopJWTTokensCleanupTicker()
}
