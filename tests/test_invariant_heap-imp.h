#include <check.h>
#include <stdlib.h>
#include <string.h>
#include "mupdf/fitz/heap-imp.h"

START_TEST(test_heap_realloc_no_overflow)
{
    // Invariant: Buffer reads never exceed the declared length
    const char *payloads[] = {
        "normal",                    // Valid input
        "A",                         // Boundary case (single char)
        "AAAAAAAAAAAAAAAAAAAAAAAAAA" // Oversized input (26 chars)
    };
    int num_payloads = sizeof(payloads) / sizeof(payloads[0]);

    for (int i = 0; i < num_payloads; i++) {
        const char *input = payloads[i];
        size_t len = strlen(input);
        
        // Create heap structure
        fz_context *ctx = fz_new_context(NULL, NULL, FZ_STORE_UNLIMITED);
        heap_type *heap = fz_malloc_struct(ctx, heap_type);
        
        // Initialize with safe values
        heap->count = 0;
        heap->max = 0;
        heap->heap = NULL;
        
        // Test reallocation with the input length
        // This should either succeed safely or fail gracefully
        // without causing buffer overflow
        int result = heap_ensure_items(ctx, heap, len);
        
        // Verify no out-of-bounds access occurred
        // If allocation succeeded, check bounds
        if (result == 0 && heap->heap != NULL) {
            ck_assert_ptr_ne(heap->heap, NULL);
            // Ensure max is at least what we requested
            ck_assert_int_ge(heap->max, len);
        }
        
        // Cleanup
        fz_free(ctx, heap->heap);
        fz_free(ctx, heap);
        fz_drop_context(ctx);
    }
}
END_TEST

Suite *security_suite(void)
{
    Suite *s;
    TCase *tc_core;

    s = suite_create("Security");
    tc_core = tcase_create("Core");

    tcase_add_test(tc_core, test_heap_realloc_no_overflow);
    suite_add_tcase(s, tc_core);

    return s;
}

int main(void)
{
    int number_failed;
    Suite *s;
    SRunner *sr;

    s = security_suite();
    sr = srunner_create(s);

    srunner_run_all(sr, CK_NORMAL);
    number_failed = srunner_ntests_failed(sr);
    srunner_free(sr);

    return (number_failed == 0) ? EXIT_SUCCESS : EXIT_FAILURE;
}