package dailylesson

// Exact authored revisions and reviewed sentence/alternative sets. Images use
// the full approved publication catalog, independently of this cloze allowlist.
var mixedRevisions = map[string]string{
	"2534b539-a80f-5001-b9a4-27a5e0679539": "445da3734da17ce98a160e46176790438f40b9b0d472f7a8e41efae2eaff4d84", // appetizer
	"b20c1447-0e63-5712-9ae7-dd4aca9aeb85": "09555ca2e69146c3c12f263b81112bca8792a808b39f9538aed4771290b28c7d", // cash
	"b548ade3-8682-521e-bf26-848526519854": "8dfe89e4fa92ac1533a3a19e125e44fb542e8fa6db67366d3ec797440d1be773", // form
	"6a61d337-80e7-5d50-bc61-8b45b6a054a1": "112a5a995ccdcaa2b3b3b231a0dcc165926b98b4732daf032eae37e0d0170a0b", // gate
	"d990ce6e-4756-52e0-96e8-9926442963f2": "068a52d16045d06ce7f8e5fe0575add338bd96df5e76aee46359ecf350d6fa7d", // key card
	"bee2408e-5991-5805-923c-6fba33ab7e7a": "5c833aeac56871b2a95ce70f21333818195df88b644d6e808335709545770e5b", // luggage
	"329c4ec9-6562-568e-9e71-9134da2f0d8d": "ae7b8d69c5c0f06a7955f25b741fb7ff6be8ceb11a0ba69d64d8d3c79ecd7c01", // menu
	"0b593408-6679-5fb8-a21c-b73a525c76e5": "58371613d889e51873a9f32f55ab342ad6c6850bfb679bdc2bb2cd92a2de42fb", // parcel
	"9aa8876d-25c6-54d1-9b65-0816e604e7a5": "f21942e74d8fef34599f85723a1310994676255cfa2a9e473c595bffc9b30951", // platform
	"74a55fe5-7736-58f0-98b8-23eef998067d": "cd0640a44faeb6c57a1c47ced53c5f2a0db273c8e435d14706c04a51571b5545", // receipt
	"538cc7f5-983c-59cc-8664-173cefca65ad": "7e5dd1adff226bc216349f5e508f6d649871ed947738f22766ca4e5c48a40f87", // refund
	"c7bbc8ae-3f64-58e9-acf0-3069e317c471": "ab59382ee12c832704dabb0decff05ce7db4ae3dc130069fe432f05d51ac8f79", // signature
	"7dee60f4-25c4-5d55-9111-dc5174dcd5b5": "df4babfd8b87ea12fc7af1c5d1927509037af75b975966dbbf539cc9e9c6db98", // size
	"58529fdd-9425-5aef-a8c7-bc1086be7080": "d20c054bc0ae66bbb38f6ea80c4904c672628e68fb0dabb79881931b58e0be40", // syllabus
	"2fd7a730-c9e9-5a7b-b094-dea868bc7cd3": "c755e94cf9524b28089fcb361ded968c9033c329807ebd849532cc28ed6bc77a", // voicemail
}

type gapSet struct {
	Example string
	IDs     []string
}

var gapSets = map[string]gapSet{
	"6a61d337-80e7-5d50-bc61-8b45b6a054a1": {Example: "Our flight leaves from gate twelve.", IDs: []string{"6a61d337-80e7-5d50-bc61-8b45b6a054a1", "58529fdd-9425-5aef-a8c7-bc1086be7080", "bee2408e-5991-5805-923c-6fba33ab7e7a", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"d990ce6e-4756-52e0-96e8-9926442963f2": {Example: "My key card stopped working.", IDs: []string{"d990ce6e-4756-52e0-96e8-9926442963f2", "58529fdd-9425-5aef-a8c7-bc1086be7080", "bee2408e-5991-5805-923c-6fba33ab7e7a", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"58529fdd-9425-5aef-a8c7-bc1086be7080": {Example: "The reading list is in the syllabus.", IDs: []string{"58529fdd-9425-5aef-a8c7-bc1086be7080", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "538cc7f5-983c-59cc-8664-173cefca65ad", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"2534b539-a80f-5001-b9a4-27a5e0679539": {Example: "We shared an appetizer before dinner.", IDs: []string{"2534b539-a80f-5001-b9a4-27a5e0679539", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "58529fdd-9425-5aef-a8c7-bc1086be7080", "d990ce6e-4756-52e0-96e8-9926442963f2"}},
	"9aa8876d-25c6-54d1-9b65-0816e604e7a5": {Example: "Our train leaves from platform three.", IDs: []string{"9aa8876d-25c6-54d1-9b65-0816e604e7a5", "58529fdd-9425-5aef-a8c7-bc1086be7080", "bee2408e-5991-5805-923c-6fba33ab7e7a", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"538cc7f5-983c-59cc-8664-173cefca65ad": {Example: "The shop sent my refund to my bank account.", IDs: []string{"538cc7f5-983c-59cc-8664-173cefca65ad", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "58529fdd-9425-5aef-a8c7-bc1086be7080", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"b20c1447-0e63-5712-9ae7-dd4aca9aeb85": {Example: "Can I pay in cash?", IDs: []string{"b20c1447-0e63-5712-9ae7-dd4aca9aeb85", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "58529fdd-9425-5aef-a8c7-bc1086be7080", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"b548ade3-8682-521e-bf26-848526519854": {Example: "Please fill in this form with your name and address.", IDs: []string{"b548ade3-8682-521e-bf26-848526519854", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "bee2408e-5991-5805-923c-6fba33ab7e7a", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"7dee60f4-25c4-5d55-9111-dc5174dcd5b5": {Example: "Do you have this shirt in a smaller size?", IDs: []string{"7dee60f4-25c4-5d55-9111-dc5174dcd5b5", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "58529fdd-9425-5aef-a8c7-bc1086be7080", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"2fd7a730-c9e9-5a7b-b094-dea868bc7cd3": {Example: "Please leave a voicemail if I do not answer.", IDs: []string{"2fd7a730-c9e9-5a7b-b094-dea868bc7cd3", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "58529fdd-9425-5aef-a8c7-bc1086be7080", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"c7bbc8ae-3f64-58e9-acf0-3069e317c471": {Example: "We need your signature at the bottom of the page.", IDs: []string{"c7bbc8ae-3f64-58e9-acf0-3069e317c471", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "bee2408e-5991-5805-923c-6fba33ab7e7a", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"74a55fe5-7736-58f0-98b8-23eef998067d": {Example: "The shop emailed me a receipt after I paid.", IDs: []string{"74a55fe5-7736-58f0-98b8-23eef998067d", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", "bee2408e-5991-5805-923c-6fba33ab7e7a", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
}
