/* 
 * Copyright 2021 Matrix Origin
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#ifndef _MO_IMPL_H_
#define _MO_IMPL_H_

#include "mo.h"

#include <stdlib.h>
#include <stdbool.h>
#include <math.h>

static const int32_t RC_SUCCESS = 0;
static const int32_t RC_INFO = 1;
static const int32_t RC_WARN = 2;

static const int32_t RC_INTERNAL_ERROR = 20101;

static const int32_t RC_DIVISION_BY_ZERO = 20200;
static const int32_t RC_OUT_OF_RANGE = 20201;
static const int32_t RC_DATA_TRUNCATED = 20202;
static const int32_t RC_INVALID_ARGUMENT = 20203;

static const int32_t LEFT_IS_SCALAR = 1;
static const int32_t RIGHT_IS_SCALAR = 2;


#include "bitmap.h"
#include "xcall.h"

#endif /* _MO_IMPL_H_ */
