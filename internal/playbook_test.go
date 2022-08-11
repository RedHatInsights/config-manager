package internal

import (
	"testing"

	"config-manager/internal/config"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGeneratePlaybook(t *testing.T) {
	config.DefaultConfig.PlaybookFiles = "./testdata/"

	tests := []struct {
		description string
		input       map[string]string
		want        string
		wantError   error
	}{
		{
			description: "all enabled",
			input: map[string]string{
				"test1": "enabled",
				"test2": "enabled",
			},
			want: `---
# Service Enablement playbook

- name: Test 1
  hosts: localhost
  vars:
    insights_signature_exclude: /hosts,/vars/insights_signature
    insights_signature: !!binary |
      TFMwdExTMUNSVWRKVGlCUVIxQWdVMGxIVGtGVVZWSkZMUzB0TFMwS1ZtVnljMmx2YmpvZ1IyNTFV
      RWNnZGpFS0NtbFJTVlpCZDFWQldVbENRbk56ZG5jMU9FUXJhalZ3VGtGUloycG1lRUZCYTFFeWRV
      RXpZV3N6WjNGcVRWUjVTbXBYUkVoNFdVaG5LeTh6Y25aalpub0tWRmRrZWtkdGVrbEZhbXRhUVV0
      RGNXdHNUWEZDWm1NelZFRXphek5tYlRGTVJITkZOa1pRTXpOTFYzbEZabFpqWVZaS2FtdFZXalJR
      SzI5RU9FWk1XQXBGZVdocFltbHRTbWx6YkVWUFpuVlZLM0ZSWW1sYU9XNVJiRFp4TjJGQ1oyNXdh
      RTlzTDFOTlZUWlRObTkzV0hjNFoxUmxPVE5vV1hkS1pIVnJhVlpMQ2tWUE1FNVBRVU5uYTFZeGEx
      bHRiVkV6TUROcWNVOXdRbXBEZVM5cFVrcFFkbGxXVVhveGFtbGthVmRLSzFSWFl6SlhZVWxZU21s
      VE9WZGpUMHhTVW5RS2VUZFRhV05XT0VOTVdsSkhVV055UnpoamFYcEZSalZRY1RFMmNtWndhVFJy
      ZFZWTFZrTlNTazVIY1RBNFdXZERNbmhhVFZFNWNYRXJOVzlLY0d0ck9RcEpkSGRPYzBKTVNDdDBS
      R3dyVkdKTE5saDNkMjFpZDBGbVNESkdiMEZTV2pGT1IxZ3JVRTFCY1cxVlNXZzFOREJCVUdRNVMy
      MHhRV1Z6TUZkNVZsb3pDalF6TVRGYVRuTnNhRzloYms1aWFuTllLMjlxTmpGSVMyUlBWamRpZUhk
      S2QybGpRM1JITldoTFFWTTVlVUZ3YVhScFVUTkNVVGxtTjJGNlNrazROR29LU1ZKMlJGSkhUM293
      V0dsVWFISlNSa01yTVdOcFQyRjRhVnBETVVaWk9UTkVRVXAwTkRkNWRuRXlSa1FyU2tKM1VqTlha
      eXRDUm5oMGRYTnlibkpyYUFwcFZWcGFXVzA1WjFjeWEzWjJVVGh2U0VKQmVHeFdkMmxSU0hOa1RU
      ZzVSVk5PTWxWclVrMVViVFJwVkd4NGRVSjVRM0IwTTB4a1MwNVJSRWgwY0hWckNtNDNWak5xVkhs
      b2MwZFpabmRHWVU4Mk9FWlNjWG94TkdKTlNVRkZRbFJhYWt0UldVazFhbk5VYWxkdlppOWpVRWxo
      V2poblJsWlRkak5pYUdNMk9IUUtVbTFsV1docU5qWk1XRFI2WVdwTU1sbExaVTF2UTJsNE9URkNZ
      emx2UlVJNGVIaDRPSFF6Y0ZKVWRWZ3JRa0p3TDNaUWNYZEthV0ZMVEZoaWExVTRaQXBZUVhWUGJr
      WTNkWGgwU1QwS1BYUm1aVE1LTFMwdExTMUZUa1FnVUVkUUlGTkpSMDVCVkZWU1JTMHRMUzB0Q2c9
      PQ==
  tasks:
    - name: Print 1 setup
      debug:
        msg: "1 setup"

- name: Test 2
  hosts: localhost
  vars:
    insights_signature_exclude: /hosts,/vars/insights_signature
    insights_signature: !!binary |
      TFMwdExTMUNSVWRKVGlCUVIxQWdVMGxIVGtGVVZWSkZMUzB0TFMwS1ZtVnljMmx2YmpvZ1IyNTFV
      RWNnZGpFS0NtbFJTVlpCZDFWQldVbENRMjFOZG5jMU9FUXJhalZ3VGtGUloyZEZVUzh2VVdkaGRI
      b3Jabk40TlRSdFVIbHRVbE5zWTIxc09GUkhNM0ZOVGt0VmJHb0thR1l4ZG5FdlRtdGtWMU5JY0VG
      VldERjRjV0VyVGxkb1VIUnBOM0JtTlZkUWNrWktkMk14WTBKb1pFWlBaa2xRTDNaNmJUWkZNRWxO
      UTJreWIxUTBjd29yTmxCWFprNW5RWEJPVjJKclRHcGhUVGc0Y1RZdllWaHhhVXMwU0drNVpXVTBP
      RzFIUmxaS2Iya3hkVUpITmpBeGJFdE5hRGh6UlU4ell6QlJNalZoQ2pScFpsTjFVVWRPU0N0c1FY
      WmxlVUpwYW13M0wxUmhiRVpWUm1Kbk1sQnRiMHc1Ym05RVJGaGhPRzlHVW1Vd1NIQnhZMGM1YXk5
      SFFXTkRjVmg1ZVhrS1RtWlZibkEwV1ZOeU9VaFBXRFp1YUZac1NsVmFRbnBqVERkaVlrSk1kRzR5
      TURkbmNuZHhhekp1ZW1SdmRGWjFXVm8zVmxoU2JrRlNjMjVSZWt0R1V3cFNRWG9yVnpkTmF5OWpP
      RW93TjJsbk4yaFVlSEl4Y0dKRVVrTnpWR3haYW1KT1IwSkxSekZ2TWtKRVJXaGpkak5OZUVZM2NH
      MXRiV0pDZHpCT1ZqTkhDblJMZFZCUVUxaFRlVzFTUmtaeFN5dHVkVmxzU1dVMlEyMDRVMnhHVUhW
      T2VtVlpXVnAyUkhobmNUTlROUzl4WldGcmMwbFNOMXByVm1GVFpHbGllRUlLYTJsU2JDOU9Wa2RX
      UzB3eVQwd3pMMnRNTHpoTk5YbE5NVzEzYWxkck5FVm1OVFZvUmtVNVVGVjZlbVJsVWpscFZFSlJW
      WFU1YWpsTmFtdGhSbWw1UndvMFNIUXpNbU0wWnpKU2NtRnVXV2xNYlc1amFqQjNjWHB2ZUVkdlZr
      RnZRbGhHTUhRNFZsVjNabVZOWjBWelJqQXpUbTk1TVUweFoycFdWRVJwUkZOb0NtTjNTSFJGZWxV
      NVNtZERhbWRZUjJobU5VWXdTWGRpYTNwU1ptTkJOall5ZDBkSVNXa3lWVlpwZERkeFZuVkRWVGd3
      V21wS2FFaDJVMUJvVW1OTFdtVUtNRkZ1UVRSamRXdDRibGRTYUhaV1VEZG5kazVYTmpGVVUxTTFk
      M0ZzWVRCWWVsZ3lTalJpUVN0blIwdE5jR2hSWldsNWVIaHVXR2RGYVVKTU1qTXJaUXBVZUZWeU1E
      VnZPSEZTT0QwS1BUQmtWalFLTFMwdExTMUZUa1FnVUVkUUlGTkpSMDVCVkZWU1JTMHRMUzB0Q2c9
      PQ==
  tasks:
    - name: Print 2 setup
      debug:
        msg: "2 setup"
`,
		},
		{
			description: "all disabled",
			input: map[string]string{
				"test1": "disabled",
				"test2": "disabled",
			},
			want: `---
# Service Enablement playbook

- name: Test 1
  hosts: localhost
  vars:
    insights_signature_exclude: /hosts,/vars/insights_signature
    insights_signature: !!binary |
      TFMwdExTMUNSVWRKVGlCUVIxQWdVMGxIVGtGVVZWSkZMUzB0TFMwS1ZtVnljMmx2YmpvZ1IyNTFV
      RWNnZGpFS0NtbFJTVlpCZDFWQldVbENRa1ZOZG5jMU9FUXJhalZ3VGtGUmFEWllkeTh2V1hSUU0y
      dEhiMU4wVW1SYVZFWmlUbk5HWmpWRlVVdzJlbll2Y2sxNU5UQUtNMDVFWjFKYVVtVk5lbVpyYUU1
      TWJWRXhRM2N3UVZKMlRrMUNOWGRFU0ZOUU0ySmtNM0I2SzNKU1RGcDZRWFZMVTNsRFEzVnhhRFZF
      Y0ZCTVZFdFpUQXBYYmxKb01ua3JNMlk0WWtWRFVXaFRRblJWYWpWemFFOVBaVWxtTVdKelNXczNj
      MlZ4TldoeVkxWXhWR3RrUWpZclJFZFFiSE5QZVRNNVNGWlhhbVJGQ25OcGQxaHRaVmRxVGtoNFMy
      azFjWFZ4U1RseGEzUkNibVZoVVhwdE4xaHNSSFZKVTB0dmR6Y3JWbTFJVFZGMVVURkhSRXRQWVdO
      d2FYSk9lRVZpWW5VS1EyWnlaM1pGUVVOUVRuSkNTRk5KTWtnd1pUSTVjR2hJTkZWQmJqQjJZa2d6
      TldjMVVqWnljV0oyTjFob1VEUjRVbFJ3WmtoR2FtOU1jbXRuYmpaUFdnbzFTVVJrTDIxb01VMVhX
      R2RyV2xKdlZGRnNkbXMyYlZwYWVXWjVTVkU1WjIxWmJHcFlRMlV3UzB4eGJXdHdTamRHTUdJNE5F
      dE5TMFZrTUdoSVVIUnlDbWhwU1docGNYSnVZM1ZTYWtOb2NEWnpWRzFaV0ZCbmVGWm1SMHhQVjNW
      aGRFTmxlazk0YmtadVZtbHFjekpLZGpGMVMweHNWMUp2VUZjek1HNHhWMklLTjAxd1MzVndSRUpz
      Y0VOUU5HOXBUMU50Tmk5SmNWUkpPVUY0TVZNck4zRkpiVkpLU2xsME1YaHdhM056VjFCc1FtUlpl
      a2cwWVVwb1NqaENOVE5SZGdwa1RWWlNkakZKYzI5UGNqUjVkMk5DYWtGQ1QxUktUVmRYZFRWWU5H
      dFFhMFZsTUVsS1IyZzRNbVZGT1dkb1MwZGxNa3dyZFV4MWRtOVdTRk5YZUZwWENtaEhNMmxaVVdo
      c05HSjFRVTkwZWtjck5rSTFUbUpWVkU1d09YZFBkMUF5UTNrM2RIQllhVmxLUW14M05DOHhRMFJw
      VEVKVmEyTjBUelJoVEhVeFpVWUtNV3BPYTAxQ1FVTm5ORGhKVWxWSlRYSXpVa3huWkROak1FTXpL
      MWgwVVd0UVdWVnZXVmxaVjIxNlRrOU9jMU5ET0hwd05uSTFOekkwZW5obVFWUk5kUXB2YURsc1RU
      ZHlWRk4xTkQwS1BWVjVjRVlLTFMwdExTMUZUa1FnVUVkUUlGTkpSMDVCVkZWU1JTMHRMUzB0Q2c9
      PQ==
  tasks:
    - name: Print 1 removed
      debug:
        msg: "1 removed"

- name: Test 2
  hosts: localhost
  vars:
    insights_signature_exclude: /hosts,/vars/insights_signature
    insights_signature: !!binary |
      TFMwdExTMUNSVWRKVGlCUVIxQWdVMGxIVGtGVVZWSkZMUzB0TFMwS1ZtVnljMmx2YmpvZ1IyNTFV
      RWNnZGpFS0NtbFJTVlpCZDFWQldVbENRalE0ZG5jMU9FUXJhalZ3VGtGUmFETklRUzh2VXpkdWVU
      aGxNbHBpVEhRMk1WQnlWMlZLU2pabGVFZFdaa2xYYzJwWkwxWUtXRFpHUmpsbk1GRlRhRlppTVRG
      bGNWVkRhRTFvUm1kRGRXWm9NSHBNVDNSdlltSm9jVTgyWmxoaVpGcFdiMFZDTVUxSWMzbGxTVWxY
      UVROaGFGWjRiUXAwUWtJdlVGTnNaMkp3VnpGek5rSmpVRlE0TkZKbVpHMDBaa3N6Vmk5Tlp6RkJa
      akJVY1VkS2IySm9XaXRQWmxweFNuRmlaWGhYVkdkM1JWbzNOMHhYQ2tVMGVFZ3llalZXTTFod1Nq
      UlhMME13YlcxMldEWmpkV0V3Y1VaT1p6VkNVR0pwZVZwWk9FaFBjblJDYTNkek9FMTBjVEZzYUU5
      UWN6aFdTRWxpYldrS1F6WXpkV05IYzFkaGFtaE1WR2R3TWpaMk1HSm1OVVZGYzA1d1pucHFVbE5t
      VW5GSGFTOXZaRWRzV0U0MlJraGpNMVZ6TnpVdlZEaFJjVGxGWjNWTVV3cFBVekp5TVhGVU9IZEhi
      MGg2V0d4cGJ6RjNWRVp0ZDJOT1NGZHlNVUV2UVRVdlprSnlaRE5qVEVWd1VFTm1abkpZWVVKQ1pV
      eERTMDlGWVhJM1lreFFDbVI1SzBsRVJXVm9RbEpzVTNsclJtSTFibWRtY1dWcFlteDFVVlJVTVdS
      MlJGQmhSV001T1VSdFkyTldNRFZhZDNsUlN6VkZObEZ0YzFkeVQxbHBUM01LTDB0dVdqSXpkakUz
      VTFKWGIzZDNLMDFwV0V0NWVVMVFjalZRTjNRNVducDBSalpUV1hVMU5uWXpjbUp3VEhWalJrcHJj
      WGQzTWpReVZFVkZRM1F5VmdwWGNYcEdVSGxQTWpCT2NtYzNSVVZNYVZwMk1uWjFPRGw1WTFGVFRG
      Sm9TVGw1V1V0NWNGTjNWRlJOYlVOclNHWTFiMVZ0UjFSS1FUUkVkR1ZIU0U0ekNrVllPRmRNTlVo
      V1VtdFdNV1pXV2xST1VqQkpVVWxIUjFGSVJXdG1SQzluZVU1aFVDdDVialEwUzJVd1JGVmhhelZR
      VTJKSFR6bHpNVmw1YTFob1QzSUtXR2xxU1RBNFRYbFZjRTF6T0d0eGJtSmFPWFE0WW1vMGNXSk9U
      V0pSVlRaSVIzTlhTRmxoUTFoSU5XeE1RWEJGVEZKc2RYcFlRMlpFYmtobksweHhkd3BKUWsxSFZt
      VkpTVThyVVQwS1BTOHlOaklLTFMwdExTMUZUa1FnVUVkUUlGTkpSMDVCVkZWU1JTMHRMUzB0Q2c9
      PQ==
  tasks:
    - name: Print 2 removed
      debug:
        msg: "2 removed"
`,
		},
		{
			description: "service without task",
			input: map[string]string{
				"test1":              "disabled",
				"test2":              "disabled",
				"serviceWithoutPlay": "disabled",
			},
			want: `---
# Service Enablement playbook

- name: Test 1
  hosts: localhost
  vars:
    insights_signature_exclude: /hosts,/vars/insights_signature
    insights_signature: !!binary |
      TFMwdExTMUNSVWRKVGlCUVIxQWdVMGxIVGtGVVZWSkZMUzB0TFMwS1ZtVnljMmx2YmpvZ1IyNTFV
      RWNnZGpFS0NtbFJTVlpCZDFWQldVbENRa1ZOZG5jMU9FUXJhalZ3VGtGUmFEWllkeTh2V1hSUU0y
      dEhiMU4wVW1SYVZFWmlUbk5HWmpWRlVVdzJlbll2Y2sxNU5UQUtNMDVFWjFKYVVtVk5lbVpyYUU1
      TWJWRXhRM2N3UVZKMlRrMUNOWGRFU0ZOUU0ySmtNM0I2SzNKU1RGcDZRWFZMVTNsRFEzVnhhRFZF
      Y0ZCTVZFdFpUQXBYYmxKb01ua3JNMlk0WWtWRFVXaFRRblJWYWpWemFFOVBaVWxtTVdKelNXczNj
      MlZ4TldoeVkxWXhWR3RrUWpZclJFZFFiSE5QZVRNNVNGWlhhbVJGQ25OcGQxaHRaVmRxVGtoNFMy
      azFjWFZ4U1RseGEzUkNibVZoVVhwdE4xaHNSSFZKVTB0dmR6Y3JWbTFJVFZGMVVURkhSRXRQWVdO
      d2FYSk9lRVZpWW5VS1EyWnlaM1pGUVVOUVRuSkNTRk5KTWtnd1pUSTVjR2hJTkZWQmJqQjJZa2d6
      TldjMVVqWnljV0oyTjFob1VEUjRVbFJ3WmtoR2FtOU1jbXRuYmpaUFdnbzFTVVJrTDIxb01VMVhX
      R2RyV2xKdlZGRnNkbXMyYlZwYWVXWjVTVkU1WjIxWmJHcFlRMlV3UzB4eGJXdHdTamRHTUdJNE5F
      dE5TMFZrTUdoSVVIUnlDbWhwU1docGNYSnVZM1ZTYWtOb2NEWnpWRzFaV0ZCbmVGWm1SMHhQVjNW
      aGRFTmxlazk0YmtadVZtbHFjekpLZGpGMVMweHNWMUp2VUZjek1HNHhWMklLTjAxd1MzVndSRUpz
      Y0VOUU5HOXBUMU50Tmk5SmNWUkpPVUY0TVZNck4zRkpiVkpLU2xsME1YaHdhM056VjFCc1FtUlpl
      a2cwWVVwb1NqaENOVE5SZGdwa1RWWlNkakZKYzI5UGNqUjVkMk5DYWtGQ1QxUktUVmRYZFRWWU5H
      dFFhMFZsTUVsS1IyZzRNbVZGT1dkb1MwZGxNa3dyZFV4MWRtOVdTRk5YZUZwWENtaEhNMmxaVVdo
      c05HSjFRVTkwZWtjck5rSTFUbUpWVkU1d09YZFBkMUF5UTNrM2RIQllhVmxLUW14M05DOHhRMFJw
      VEVKVmEyTjBUelJoVEhVeFpVWUtNV3BPYTAxQ1FVTm5ORGhKVWxWSlRYSXpVa3huWkROak1FTXpL
      MWgwVVd0UVdWVnZXVmxaVjIxNlRrOU9jMU5ET0hwd05uSTFOekkwZW5obVFWUk5kUXB2YURsc1RU
      ZHlWRk4xTkQwS1BWVjVjRVlLTFMwdExTMUZUa1FnVUVkUUlGTkpSMDVCVkZWU1JTMHRMUzB0Q2c9
      PQ==
  tasks:
    - name: Print 1 removed
      debug:
        msg: "1 removed"

- name: Test 2
  hosts: localhost
  vars:
    insights_signature_exclude: /hosts,/vars/insights_signature
    insights_signature: !!binary |
      TFMwdExTMUNSVWRKVGlCUVIxQWdVMGxIVGtGVVZWSkZMUzB0TFMwS1ZtVnljMmx2YmpvZ1IyNTFV
      RWNnZGpFS0NtbFJTVlpCZDFWQldVbENRalE0ZG5jMU9FUXJhalZ3VGtGUmFETklRUzh2VXpkdWVU
      aGxNbHBpVEhRMk1WQnlWMlZLU2pabGVFZFdaa2xYYzJwWkwxWUtXRFpHUmpsbk1GRlRhRlppTVRG
      bGNWVkRhRTFvUm1kRGRXWm9NSHBNVDNSdlltSm9jVTgyWmxoaVpGcFdiMFZDTVUxSWMzbGxTVWxY
      UVROaGFGWjRiUXAwUWtJdlVGTnNaMkp3VnpGek5rSmpVRlE0TkZKbVpHMDBaa3N6Vmk5Tlp6RkJa
      akJVY1VkS2IySm9XaXRQWmxweFNuRmlaWGhYVkdkM1JWbzNOMHhYQ2tVMGVFZ3llalZXTTFod1Nq
      UlhMME13YlcxMldEWmpkV0V3Y1VaT1p6VkNVR0pwZVZwWk9FaFBjblJDYTNkek9FMTBjVEZzYUU5
      UWN6aFdTRWxpYldrS1F6WXpkV05IYzFkaGFtaE1WR2R3TWpaMk1HSm1OVVZGYzA1d1pucHFVbE5t
      VW5GSGFTOXZaRWRzV0U0MlJraGpNMVZ6TnpVdlZEaFJjVGxGWjNWTVV3cFBVekp5TVhGVU9IZEhi
      MGg2V0d4cGJ6RjNWRVp0ZDJOT1NGZHlNVUV2UVRVdlprSnlaRE5qVEVWd1VFTm1abkpZWVVKQ1pV
      eERTMDlGWVhJM1lreFFDbVI1SzBsRVJXVm9RbEpzVTNsclJtSTFibWRtY1dWcFlteDFVVlJVTVdS
      MlJGQmhSV001T1VSdFkyTldNRFZhZDNsUlN6VkZObEZ0YzFkeVQxbHBUM01LTDB0dVdqSXpkakUz
      VTFKWGIzZDNLMDFwV0V0NWVVMVFjalZRTjNRNVducDBSalpUV1hVMU5uWXpjbUp3VEhWalJrcHJj
      WGQzTWpReVZFVkZRM1F5VmdwWGNYcEdVSGxQTWpCT2NtYzNSVVZNYVZwMk1uWjFPRGw1WTFGVFRG
      Sm9TVGw1V1V0NWNGTjNWRlJOYlVOclNHWTFiMVZ0UjFSS1FUUkVkR1ZIU0U0ekNrVllPRmRNTlVo
      V1VtdFdNV1pXV2xST1VqQkpVVWxIUjFGSVJXdG1SQzluZVU1aFVDdDVialEwUzJVd1JGVmhhelZR
      VTJKSFR6bHpNVmw1YTFob1QzSUtXR2xxU1RBNFRYbFZjRTF6T0d0eGJtSmFPWFE0WW1vMGNXSk9U
      V0pSVlRaSVIzTlhTRmxoUTFoSU5XeE1RWEJGVEZKc2RYcFlRMlpFYmtobksweHhkd3BKUWsxSFZt
      VkpTVThyVVQwS1BTOHlOaklLTFMwdExTMUZUa1FnVUVkUUlGTkpSMDVCVkZWU1JTMHRMUzB0Q2c9
      PQ==
  tasks:
    - name: Print 2 removed
      debug:
        msg: "2 removed"
`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got, err := GeneratePlaybook(test.input)

			if test.wantError != nil {
				if !cmp.Equal(err, test.wantError, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", err, test.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, test.want) {
					t.Errorf("%v", cmp.Diff(got, test.want))
				}
			}
		})
	}
}

func TestVerifyStatePayload(t *testing.T) {
	tests := []struct {
		desc  string
		input struct {
			current map[string]string
			payload map[string]string
		}
		want      bool
		wantError error
	}{
		{
			desc: "valid payload",
			input: struct {
				current map[string]string
				payload map[string]string
			}{
				current: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
				payload: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "disabled",
				},
			},
			want: false,
		},
		{
			desc: "payload equal to current state",
			input: struct {
				current map[string]string
				payload map[string]string
			}{
				current: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
				payload: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
			},
			want: true,
		},
		{
			desc: "additional services enabled when insights is disabled",
			input: struct {
				current map[string]string
				payload map[string]string
			}{
				current: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
				payload: map[string]string{
					"insights":            "disabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
			},
			want:      false,
			wantError: cmpopts.AnyError,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := VerifyStatePayload(test.input.current, test.input.payload)
			if test.wantError != nil {
				if !cmp.Equal(test.wantError, err, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", test.wantError, err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if test.want != got {
					t.Errorf("%v != %v", test.want, got)
				}
			}
		})
	}
}
